package models

type RespostaFormulario struct {
	ID    uint64 `gorm:"primaryKey"`
	Valor string `gorm:"type:text"`

	PedidoAdocaoID uint64       `gorm:"not null;index"`
	PedidoAdocao   PedidoAdocao `gorm:"foreignKey:PedidoAdocaoID;constraint:OnDelete:CASCADE"`

	CampoFormularioID uint64          `gorm:"not null"`
	CampoFormulario   CampoFormulario `gorm:"foreignKey:CampoFormularioID;constraint:OnDelete:CASCADE"`
}
