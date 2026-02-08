package models

type Ong struct {
	ID        uint64 `gorm:"primaryKey"`
	Nome      string `gorm:"size:150;not null"`
	Descricao string `gorm:"type:text"`

	Pets              []Pet
	PedidosAdocao     []PedidoAdocao
	FormulariosModelo []FormularioModelo
}
