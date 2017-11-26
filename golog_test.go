package main

import (
	"testing"
	"golog/log"
)

func BenchmarkGoLogTextNegative(b *testing.B) {
	log.SetLevel(log.LvERROR)
	logger := log.GetLogger("logs/app")
	defer log.DeferLogger()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.LInfo("The quick brown fox jumps over the lazy dog")
		}
	})

}

func BenchmarkGoLogTextPositive(b *testing.B) {
	log.SetLevel(log.LvINFO)
	logger := log.GetLogger("logs/app")
	defer log.DeferLogger()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.LInfo("The quick brown fox jumps over the lazy dog")
		}
	})

}
