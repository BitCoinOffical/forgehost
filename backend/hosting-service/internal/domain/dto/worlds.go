package dto

type UpsertMinecraftServerDTO struct {
	Name    string `json:"name" binding:"required"`
	Address string `json:"address" binding:"required"`
	Port    int    `json:"port" binding:"required,min=1,max=65535"`
	Edition string `json:"edition" binding:"required,oneof=java bedrock"`
}
