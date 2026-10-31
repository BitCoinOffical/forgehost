package models

import "time"

type MinecraftServer struct {
	ID        string    `db:"id" json:"id"`
	UserID    string    `db:"user_id" json:"-"`
	Name      string    `db:"name" json:"name"`
	Address   string    `db:"address" json:"address"`
	Port      int       `db:"port" json:"port"`
	Edition   string    `db:"edition" json:"edition"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`

	//cache
	Online        bool          `json:"online"`
	Latency       time.Duration `json:"latency_ms" swaggertype:"integer" format:"int64"`
	PlayersOnline int64         `json:"players_online"`
	PlayersMax    int64         `json:"players_max"`
	Version       string        `json:"version"`
	MOTD          string        `json:"motd"`
	CheckedAt     time.Time     `json:"checked_at"`
}

type MinecraftServerCheckPayload struct {
	ID      string `json:"id"`
	Address string `json:"address"`
	Port    int    `json:"port"`
	Edition string `json:"edition"`
}

type MinecraftStatus struct {
	Online        bool          `json:"online"`
	Latency       time.Duration `json:"latency_ms" swaggertype:"integer" format:"int64"`
	PlayersOnline int64         `json:"players_online"`
	PlayersMax    int64         `json:"players_max"`
	Version       string        `json:"version"`
	MOTD          string        `json:"motd"`
}
