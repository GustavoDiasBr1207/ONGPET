package utils

import (
	"fmt"
	"log"
	"time"

	"ongpet/database"
	"ongpet/models"

	"github.com/google/uuid"
)

// StartAcompanhamentoReminderScheduler inicia o agendador que roda diariamente
// às 19:38 BRT e envia lembretes para acompanhamentos com proxima_data
// até amanhã (inclusive atrasados) que ainda não receberam lembrete.
//
// Chame uma única vez em main.go após inicializar banco e serviços:
//
//	utils.StartAcompanhamentoReminderScheduler()
func StartAcompanhamentoReminderScheduler() {
	go func() {
		log.Println("📅 Agendador de lembretes de acompanhamento iniciado")

		for {
			nextRun := nextDailyRun(20, 00, -3) // 20:00 BRT (UTC-3)
			log.Printf("⏰ Próximo disparo do lembrete de acompanhamento: %s\n", nextRun.Format("02/01/2006 15:04:05"))
			time.Sleep(time.Until(nextRun))

			log.Println("🔔 Executando verificação de acompanhamentos...")
			if err := runAcompanhamentoReminders(); err != nil {
				log.Printf("⚠️ Erro ao executar lembretes de acompanhamento: %v\n", err)
			}
		}
	}()
}

// nextDailyRun calcula o próximo instante em que o relógio marcar hour:minute
// no fuso offsetHours (ex: -3 para BRT). Se esse momento já passou hoje, retorna amanhã.
func nextDailyRun(hour, minute, offsetHours int) time.Time {
	loc := time.FixedZone("custom", offsetHours*3600)
	now := time.Now().In(loc)

	next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, loc)
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return next
}

// runAcompanhamentoReminders busca todos os acompanhamentos ativos onde:
// - lembrete_enviado = false (nunca enviado ou resetado por nova proxima_data)
// - proxima_data <= fim de amanhã (atrasados + vencendo amanhã)
// Após enviar, marca lembrete_enviado = true para não reenviar.
func runAcompanhamentoReminders() error {
	db := database.GetDB()

	// Fim de amanhã em UTC — inclui tudo que vence até o final do dia de amanhã
	now := time.Now().UTC()
	tomorrow := now.AddDate(0, 0, 1)
	endOfTomorrow := time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 59, 0, time.UTC)

	var acompanhamentos []models.Acompanhamento
	if err := db.
		Preload("Pet").
		Where("status = ? AND lembrete_enviado = ? AND proxima_data <= ?",
			models.AcompanhamentoAtivo, false, endOfTomorrow).
		Find(&acompanhamentos).Error; err != nil {
		return fmt.Errorf("erro ao buscar acompanhamentos: %w", err)
	}

	if len(acompanhamentos) == 0 {
		log.Println("✅ Nenhum lembrete pendente para enviar")
		return nil
	}

	log.Printf("📋 %d lembrete(s) pendente(s) para enviar\n", len(acompanhamentos))

	for _, a := range acompanhamentos {
		a := a // captura para goroutine

		var ong models.Ong
		if err := db.First(&ong, "id = ?", a.OngID).Error; err != nil {
			log.Printf("⚠️ ONG não encontrada para acompanhamento %s: %v\n", a.ID, err)
			continue
		}

		// ── Pet nome ────────────────────────────────────────────────────────
		petNome := "Pet"
		if a.Pet.ID != uuid.Nil {
			petNome = a.Pet.Nome
		}

		// ── Próxima data formatada ───────────────────────────────────────────
		proximaDataFmt := ""
		if a.ProximaData != nil {
			proximaDataFmt = a.ProximaData.Format("02/01/2006")
		}

		frequenciaStr := string(a.Frequencia)

		// Marca lembrete_enviado = true ANTES de enviar para evitar reenvio
		// mesmo se o envio falhar parcialmente (email ok, whatsapp falhou)
		if err := db.Model(&models.Acompanhamento{}).
			Where("id = ?", a.ID).
			Update("lembrete_enviado", true).Error; err != nil {
			log.Printf("⚠️ Falha ao marcar lembrete_enviado (id: %s): %v\n", a.ID, err)
			continue // não envia se não conseguiu marcar, evita duplicatas
		}

		// ─── Email ───────────────────────────────────────────────────────────
		go func() {
			mailer := GetMailer()
			if mailer == nil {
				log.Println("⚠️ Mailer indisponível, email de lembrete não enviado")
				return
			}

			subject, body := NewAcompanhamentoReminderEmail(
				petNome,
				a.AdotanteNome,
				a.AdotanteTelefone,
				a.AdotanteEmail,
				proximaDataFmt,
				frequenciaStr,
			)

			if err := mailer.Send([]string{ong.Email}, subject, body); err != nil {
				log.Printf("⚠️ Erro ao enviar email de lembrete (ong: %s, pet: %s): %v\n",
					ong.Email, petNome, err)
			} else {
				log.Printf("✅ Email de lembrete enviado → %s (pet: %s, adotante: %s)\n",
					ong.Email, petNome, a.AdotanteNome)
			}
		}()

		// ─── WhatsApp ────────────────────────────────────────────────────────
		if ong.Telefone != "" {
			SendWhatsAppAcompanhamentoReminder(
				ong.Telefone,
				petNome,
				a.AdotanteNome,
				a.AdotanteTelefone,
				a.AdotanteEmail,
				proximaDataFmt,
				frequenciaStr,
			)
			log.Printf("✅ WhatsApp de lembrete enviado → %s (pet: %s, adotante: %s)\n",
				maskPhoneNumber(ong.Telefone), petNome, a.AdotanteNome)
		} else {
			log.Printf("⚠️ Telefone da ONG não configurado (id: %s), WhatsApp não enviado\n", ong.ID)
		}
	}

	return nil
}