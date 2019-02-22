package main

import (
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/mattermost/mattermost-server/model"
	"github.com/mattermost/mattermost-server/plugin"
)

var schema1 = `
CREATE TABLE IF NOT EXISTS PushNotifications (
    id CHAR(26),
	platform TEXT,
	server_id TEXT,
	device_id TEXT,
	category TEXT,
	sound TEXT,
	message TEXT,
	badge INT,
	content_available INT,
	team_id CHAR(26),
	channel_id CHAR(26),
	post_id CHAR(26),
	root_id CHAR(26),
	channel_name TEXT,
	type TEXT,
	sender_id TEXT,
	override_username TEXT,
	override_icon_url TEXT,
	from_webhook TEXT,
	version TEXT,
	sent DATETIME,
	ack DATETIME
);`

var schema2 = `
CREATE TABLE IF NOT EXISTS EnqueuedPushNotifications (
    id CHAR(26),
    type TEXT,
    user_id CHAR(26),
    channel_id CHAR(26),
    post_id CHAR(26),
	enqueued DATETIME
);`

// Plugin the main struct for everything
type Plugin struct {
	plugin.MattermostPlugin
	db                *sqlx.DB
	configurationLock sync.RWMutex
	configuration     *configuration
}

func (p *Plugin) OnActivate() error {
	config := p.getConfiguration()
	db, err := sqlx.Open(config.DBDriver, config.DBConn)
	if err != nil {
		return err
	}
	p.db = db
	_, err = db.Exec(schema1)
	if err != nil {
		return err
	}
	_, err = db.Exec(schema2)
	if err != nil {
		return err
	}
	return nil
}

func (p *Plugin) Implements(hookId int) bool {
	return hookId == plugin.PushNotificationWillBeSentId || hookId == plugin.PushNotificationHasBeenSentId || hookId == plugin.PushNotificationEnqueuedId
}

func (p *Plugin) PushNotificationWillBeSent(c *plugin.Context, notification *model.PushNotification) *model.PushNotification {
	_, err := p.db.Exec(
		"INSERT INTO PushNotifications VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NULL)",
		notification.Id,
		notification.Platform,
		notification.ServerId,
		notification.DeviceId,
		notification.Category,
		notification.Sound,
		notification.Message,
		notification.Badge,
		notification.ContentAvailable,
		notification.TeamId,
		notification.ChannelId,
		notification.PostId,
		notification.RootId,
		notification.ChannelName,
		notification.Type,
		notification.SenderId,
		notification.OverrideUsername,
		notification.OverrideIconUrl,
		notification.FromWebhook,
		notification.Version,
	)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return notification
}

func (p *Plugin) PushNotificationHasBeenSent(c *plugin.Context, notification *model.PushNotification) {
	_, err := p.db.Exec(
		"UPDATE PushNotifications SET ack=NOW() WHERE server_id=? AND device_id=? AND post_id=? AND sender_id=?",
		notification.ServerId,
		notification.DeviceId,
		notification.PostId,
		notification.SenderId,
	)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

func (p *Plugin) PushNotificationEnqueued(c *plugin.Context, notificationId, notificationType, userId, channelId, postId string) {
	_, err := p.db.Exec(
		"INSERT INTO EnqueuedPushNotifications VALUES (?, ?, ?, ?, ?, NOW())",
		notificationId,
		notificationType,
		userId,
		channelId,
		postId,
	)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}
