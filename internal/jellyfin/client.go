package jellyfin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Stoufiler/JellyWolProxy/internal/config"
	"github.com/sirupsen/logrus"
)

// Client is a Jellyfin API client
type Client struct {
	config config.Config
	logger *logrus.Logger
	client *http.Client
}

// NewClient creates a new Jellyfin client
func NewClient(cfg config.Config, logger *logrus.Logger) *Client {
	return &Client{
		config: cfg,
		logger: logger,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

// Session represents an active Jellyfin session
type Session struct {
	ID               string    `json:"Id"`
	UserID           string    `json:"UserId"`
	UserName         string    `json:"UserName"`
	Client           string    `json:"Client"`
	DeviceName       string    `json:"DeviceName"`
	PlayState        PlayState `json:"PlayState"`
	NowPlayingItem   *Item     `json:"NowPlayingItem,omitempty"`
	LastActivityDate time.Time `json:"LastActivityDate"`
}

// PlayState represents playback state
type PlayState struct {
	IsPaused      bool   `json:"IsPaused"`
	PositionTicks int64  `json:"PositionTicks"`
	CanSeek       bool   `json:"CanSeek"`
	PlayMethod    string `json:"PlayMethod"`
}

// Item represents a media item
type Item struct {
	Name              string            `json:"Name"`
	Type              string            `json:"Type"`
	SeriesName        string            `json:"SeriesName,omitempty"`
	SeasonName        string            `json:"SeasonName,omitempty"`
	IndexNumber       int               `json:"IndexNumber,omitempty"`
	ParentIndexNumber int               `json:"ParentIndexNumber,omitempty"`
	RunTimeTicks      int64             `json:"RunTimeTicks"`
	PremiereDate      string            `json:"PremiereDate,omitempty"`
	ProductionYear    int               `json:"ProductionYear,omitempty"`
	ImageTags         map[string]string `json:"ImageTags,omitempty"`
}

// GetActiveSessions retrieves all active sessions from Jellyfin
func (c *Client) GetActiveSessions() ([]Session, error) {
	if c.config.JellyfinUrl == "" || c.config.ApiKey == "" {
		return nil, fmt.Errorf("jellyfin not configured")
	}

	url := fmt.Sprintf("http://%s:%d/Sessions", c.config.ForwardIp, c.config.ForwardPort)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Emby-Token", c.config.ApiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("jellyfin API returned %d: %s", resp.StatusCode, string(body))
	}

	var sessions []Session
	if err := json.NewDecoder(resp.Body).Decode(&sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

// FormatDuration converts ticks to a readable duration
func FormatDuration(ticks int64) string {
	duration := time.Duration(ticks * 100) // Ticks are in 100-nanosecond units
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	return fmt.Sprintf("%dm %ds", minutes, seconds)
}

// GetProgress returns playback progress as percentage
func (s *Session) GetProgress() float64 {
	if s.NowPlayingItem == nil || s.NowPlayingItem.RunTimeTicks == 0 {
		return 0
	}
	return float64(s.PlayState.PositionTicks) / float64(s.NowPlayingItem.RunTimeTicks) * 100
}

// GetItemTitle returns formatted item title
func (s *Session) GetItemTitle() string {
	if s.NowPlayingItem == nil {
		return ""
	}

	item := s.NowPlayingItem
	if item.Type == "Episode" {
		if item.SeriesName != "" {
			return fmt.Sprintf("%s - S%02dE%02d - %s",
				item.SeriesName,
				item.ParentIndexNumber,
				item.IndexNumber,
				item.Name)
		}
	}

	return item.Name
}
