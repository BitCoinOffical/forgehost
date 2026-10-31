package checker

import (
	"context"
	"fmt"
	"time"

	"github.com/BitCoinOffical/forgehost/hosting-service/internal/domain/models"
	minecraftstatus "github.com/mcstatus-io/mcutil/v4/status"
)

type MinecraftChecker struct {
	timeout time.Duration
}

func NewMinecraftChecker(timeout time.Duration) *MinecraftChecker {
	return &MinecraftChecker{
		timeout: timeout,
	}
}

func (m *MinecraftChecker) Check(ctx context.Context, world *models.MinecraftServer) (*models.MinecraftStatus, error) {
	if world.Port < 1 || world.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", world.Port)
	}

	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	started := time.Now()

	result := &models.MinecraftStatus{
		Online: true,
	}

	switch world.Edition {
	case "java":
		resp, err := minecraftstatus.Modern(
			ctx,
			world.Address,
			uint16(world.Port),
		)
		if err != nil {
			return nil, fmt.Errorf("minecraftstatus.Modern: %w", err)
		}

		result.Version = resp.Version.Name.Clean
		result.MOTD = resp.MOTD.Clean

		if resp.Players.Online != nil {
			result.PlayersOnline = *resp.Players.Online
		}

		if resp.Players.Max != nil {
			result.PlayersMax = *resp.Players.Max
		}

	case "bedrock":
		resp, err := minecraftstatus.Bedrock(
			ctx,
			world.Address,
			uint16(world.Port),
		)
		if err != nil {
			return nil, fmt.Errorf("minecraftstatus.Bedrock: %w", err)
		}

		if resp.Version != nil {
			result.Version = *resp.Version
		}

		if resp.MOTD != nil {
			result.MOTD = resp.MOTD.Clean
		}

		if resp.OnlinePlayers != nil {
			result.PlayersOnline = *resp.OnlinePlayers
		}

		if resp.MaxPlayers != nil {
			result.PlayersMax = *resp.MaxPlayers
		}

	default:
		return nil, fmt.Errorf("unsupported edition: %s", world.Edition)
	}

	result.Latency = time.Since(started)

	return result, nil
}
