package tracker

import "sync"

var (
	TrackedCoins = make(map[string]bool) // Глобальная мапа для отслеживаемых криптовалют
	TrackedMutex sync.Mutex              // Мьютекс для синхронизации доступа к мапе

	StopChannels = make(map[string]chan struct{}) // Глобальная мапа для каналов остановки
	StopMutex    sync.Mutex                       // Мьютекс для синхронизации доступа к мапе
)
