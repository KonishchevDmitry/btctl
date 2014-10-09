// Decision-making center.
package dmc

import (
	"log"
	"time"
	"btctl/ipt"
)

const NetworkUsageCollectionPeriod time.Duration = 5 * time.Second
const packetsSpeedThreshold = 100
const bytesSpeedThreshold = 2000
const moratoriumTime = 5 * time.Minute
const noUsageMoratoriumTime = 1 * time.Minute


type Dmc struct {
	lastStat ipt.NetworkUsage
	lastStatTime time.Time
	moratoriumTill time.Time
	noUsageSince time.Time
}

// TODO: error?
func (dmc *Dmc) OnNetworkUsageStat(statTime time.Time, stat ipt.NetworkUsage) error {
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

	return nil
}

func (dmc *Dmc) onNetworkUsageStat(statTime time.Time, stat ipt.NetworkUsage) bool {
	packets := int64(stat.Packets) - int64(dmc.lastStat.Packets)
	bytes := int64(stat.Bytes) - int64(dmc.lastStat.Bytes)

	if packets < 0 || bytes < 0 {
		log.Println("iptables counters reset detected.")
		return false
	}

	period := statTime.Sub(dmc.lastStatTime)
	if period < NetworkUsageCollectionPeriod / 2 {
		log.Println("Clock screw detected!")
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
			log.Printf("Turn on the moratorium. Packets speed: %d, bytes speed: %d.",
				uint64(packetsSpeed), uint64(bytesSpeed))
		}

		dmc.moratoriumTill = statTime.Add(moratoriumTime)
	}

	log.Println(statTime, period, packets, bytes, packetsSpeed, bytesSpeed)

	return true
}

func (dmc *Dmc) turnOffMoratorium(reason string) {
	log.Printf("Turn off the moratorium: %s.", reason)
	dmc.moratoriumTill = *new(time.Time)
	dmc.noUsageSince = *new(time.Time)
}
