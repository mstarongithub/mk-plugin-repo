package util

import "github.com/rs/zerolog/log"

type ZerologWrapper struct{}

func (z *ZerologWrapper) Errorf(format string, args ...any) {
	log.Error().Msgf(format, args...)
}

func (z *ZerologWrapper) Debugf(format string, args ...any) {
	log.Debug().Msgf(format, args...)
}

func (z *ZerologWrapper) Infof(format string, args ...any) {
	log.Info().Msgf(format, args...)
}

func (z *ZerologWrapper) Warnf(format string, args ...any) {
	log.Warn().Msgf(format, args...)
}
