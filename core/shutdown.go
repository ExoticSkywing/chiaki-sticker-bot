package core

import (
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	tele "gopkg.in/telebot.v3"
)

const shutdownRetryMessage = "Bot is restarting for a deployment, so your current task was interrupted.\n" +
	"Please retry with /import after a moment.\n\n" +
	"機器正在更新重啟，這次轉檔已中斷。\n" +
	"請稍後重新使用 /import。"

type shutdownSession struct {
	uid int64
	ud  *UserData
}

func failActiveSessionsForShutdown() {
	sessions := snapshotActiveSessions()
	log.WithField("sessions", len(sessions)).Info("Failing active sessions for shutdown.")

	var notifyWg sync.WaitGroup
	for _, s := range sessions {
		if s.ud.cancel != nil {
			s.ud.cancel()
		}
		recordShutdownSessionFailure(s)
		notifyWg.Add(1)
		go func(session shutdownSession) {
			defer notifyWg.Done()
			notifyShutdownSessionFailure(session)
		}(s)
	}

	done := make(chan struct{})
	go func() {
		notifyWg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		log.Warn("Timed out notifying active sessions about shutdown interruption.")
	}

	for _, s := range sessions {
		cleanUserData(s.uid)
	}
}

func snapshotActiveSessions() []shutdownSession {
	users.mu.Lock()
	defer users.mu.Unlock()

	sessions := make([]shutdownSession, 0, len(users.data))
	for uid, ud := range users.data {
		sessions = append(sessions, shutdownSession{
			uid: uid,
			ud:  ud,
		})
	}
	return sessions
}

func recordShutdownSessionFailure(s shutdownSession) {
	action := s.ud.command
	packID := s.ud.stickerData.id
	if s.ud.command == "import" && s.ud.lineData != nil {
		action = "import_" + s.ud.lineData.Store
		packID = s.ud.lineData.Id
	}
	if strings.TrimSpace(action) == "" {
		action = "session"
	}
	insertShutdownEvent(s.uid, action, packID, "fail: deployment interrupted, retry required")
}

func insertShutdownEvent(userID int64, action string, packID string, status string) {
	if db == nil {
		return
	}
	_, err := db.Exec(
		"INSERT INTO events (user_id, action, pack_id, status) VALUES (?, ?, ?, ?)",
		userID, action, packID, status,
	)
	if err != nil {
		log.Debugln("insertShutdownEvent error:", err)
	}
}

func notifyShutdownSessionFailure(s shutdownSession) {
	if b == nil {
		return
	}
	_, err := b.Send(&tele.User{ID: s.uid}, shutdownRetryMessage)
	if err != nil {
		log.WithError(err).WithField("uid", s.uid).Warn("Failed to notify user about shutdown interruption.")
	}
}
