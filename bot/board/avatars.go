package board

import (
	"context"
	"image"
	"image/png"
	"net/http"
	"sync"
	"time"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"golang.org/x/image/draw"
)

const avatarCacheTTL = 10 * time.Minute

type cachedAvatar struct {
	img       image.Image
	expiresAt time.Time
}

var (
	avatarMu    sync.RWMutex
	avatarCache = make(map[string]cachedAvatar)
	httpClient  = &http.Client{Timeout: 4 * time.Second}
)

func FetchAvatars(ctx context.Context, client *bot.Client, guildID snowflake.ID, userIDs []string) map[string]image.Image {
	result := make(map[string]image.Image, len(userIDs))
	if client == nil || len(userIDs) == 0 {
		return result
	}

	now := time.Now()
	var missing []string

	avatarMu.RLock()
	for _, id := range userIDs {
		if entry, ok := avatarCache[id]; ok && entry.expiresAt.After(now) {
			if entry.img != nil {
				result[id] = entry.img
			}
		} else {
			missing = append(missing, id)
		}
	}
	avatarMu.RUnlock()

	if len(missing) == 0 {
		return result
	}

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		sem = make(chan struct{}, 6)
	)

	for _, id := range missing {
		wg.Add(1)
		go func(userID string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			img := fetchSingleAvatar(ctx, client, guildID, userID)

			avatarMu.Lock()
			avatarCache[userID] = cachedAvatar{
				img:       img,
				expiresAt: time.Now().Add(avatarCacheTTL),
			}
			avatarMu.Unlock()

			if img != nil {
				mu.Lock()
				result[userID] = img
				mu.Unlock()
			}
		}(id)
	}
	wg.Wait()

	return result
}

func fetchSingleAvatar(ctx context.Context, client *bot.Client, guildID snowflake.ID, userID string) image.Image {
	uid, err := snowflake.Parse(userID)
	if err != nil {
		return nil
	}

	var avatarURL string
	if guildID != 0 {
		if member, ok := client.Caches.Member(guildID, uid); ok {
			avatarURL = member.EffectiveAvatarURL(discord.WithFormat(discord.FileFormatPNG), discord.WithSize(128))
		}
	}

	if avatarURL == "" {
		reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		u, err := client.Rest.GetUser(uid, rest.WithCtx(reqCtx))
		if err != nil || u == nil {
			return nil
		}
		avatarURL = u.EffectiveAvatarURL(discord.WithFormat(discord.FileFormatPNG), discord.WithSize(128))
	}

	if avatarURL == "" {
		return nil
	}

	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, avatarURL, nil)
	if err != nil {
		return nil
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	src, err := png.Decode(resp.Body)
	if err != nil {
		return nil
	}

	dst := image.NewRGBA(image.Rect(0, 0, 96, 96))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}
