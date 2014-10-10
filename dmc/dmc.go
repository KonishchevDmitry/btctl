// Decision-Making Center.
package dmc

import (
	"btctl/ipt"
	"btctl/util"
	"time"
)

const NetworkUsageCollectionPeriod time.Duration = 5 * time.Second
const packetsSpeedThreshold = 100
const bytesSpeedThreshold = 2000
const moratoriumTime = 5 * time.Minute
const noUsageMoratoriumTime = 1 * time.Minute

var log = util.MustGetLogger("dmc")

type Dmc struct {
	lastStat ipt.NetworkUsage
	lastStatTime time.Time
	moratoriumTill time.Time
	noUsageSince time.Time
}

func (dmc *Dmc) OnNetworkUsageStat(statTime time.Time, stat ipt.NetworkUsage) {
	var decisionMade = false

	if ! dmc.lastStatTime.IsZero() {
		decisionMade = dmc.onNetworkUsageStat(statTime, stat)
	}

	if !dmc.moratoriumTill.IsZero() {
		// Turn off moratorium if it's expired
		if decisionMade && dmc.moratoriumTill.Unix() <= statTime.Unix() {
			dmc.turnOffMoratorium("moratorium time has expired")
		} else if !dmc.noUsageSince.IsZero() && statTime.Sub(dmc.noUsageSince) >= noUsageMoratoriumTime {
			dmc.turnOffMoratorium("zero network usage")
		}
	}

	dmc.lastStat = stat
	dmc.lastStatTime = statTime
}

func (dmc *Dmc) onNetworkUsageStat(statTime time.Time, stat ipt.NetworkUsage) bool {
	packets := int64(stat.Packets) - int64(dmc.lastStat.Packets)
	bytes := int64(stat.Bytes) - int64(dmc.lastStat.Bytes)

	if packets < 0 || bytes < 0 {
		log.Debug("iptables counters reset detected.")
		return false
	}

	period := statTime.Sub(dmc.lastStatTime)
	if period < NetworkUsageCollectionPeriod / 2 {
		log.Error("Clock screw detected!")
		return false
	}

	var floatPeriod = float64(period * 10 / time.Second) / 10
	packetsSpeed := float64(packets) / floatPeriod
	bytesSpeed := float64(bytes) / floatPeriod

	// Count zero usage time if we're under moratorium
	if !dmc.moratoriumTill.IsZero() {
		if packetsSpeed == 0 && bytesSpeed == 0 {
			if dmc.noUsageSince.IsZero() {
				dmc.noUsageSince = statTime
			}
		} else {
			if !dmc.noUsageSince.IsZero() {
				dmc.noUsageSince = *new(time.Time)
			}
		}
	}

	if packetsSpeed >= packetsSpeedThreshold || bytesSpeed >= bytesSpeedThreshold {
		if dmc.moratoriumTill.IsZero() {
			log.Info("Turn on the moratorium. Packets speed: %d, bytes speed: %d.",
				uint64(packetsSpeed), uint64(bytesSpeed))
		}

		dmc.moratoriumTill = statTime.Add(moratoriumTime)
	}

	// TODO
	log.Debug("Usage: %s %d/%d %.1f/%.1f", statTime.Format("2006.01.02 15:04:05"),
		packets, bytes, packetsSpeed, bytesSpeed)

	return true
}

func (dmc *Dmc) turnOffMoratorium(reason string) {
	log.Info("Turn off the moratorium: %s.", reason)
	dmc.moratoriumTill = *new(time.Time)
	dmc.noUsageSince = *new(time.Time)
}
