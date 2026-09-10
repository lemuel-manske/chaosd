package application

import (
	"chaosd/cli/internal/network"
	"chaosd/cli/internal/session"
)

func TranslateType(faultType session.FaultType) network.NetworkFaultType {
	switch faultType {
	case session.PartitionFaultType:
		return network.NetworkPartitionFaultType
	case session.DelayFaultType:
		return network.NetworkDelayFaultType
	default:
		panic("unknown fault type")
	}
}

func DisassembleEffects(effects []session.FaultEffect) []network.AppliedEffect {
	var disassembled []network.AppliedEffect

	for _, effect := range effects {
		disassembled = append(disassembled, network.AppliedEffect{
			NetworkName: effect.NetworkName,
			SourceIP:    effect.SourceIP,
			TargetIP:    effect.TargetIP,
		})
	}

	return disassembled
}

func AssembleEffects(effects []network.AppliedEffect) []session.FaultEffect {
	var translated []session.FaultEffect

	for _, effect := range effects {
		translated = append(translated, session.FaultEffect{
			NetworkName: effect.NetworkName,
			SourceIP:    effect.SourceIP,
			TargetIP:    effect.TargetIP,
		})
	}

	return translated
}
