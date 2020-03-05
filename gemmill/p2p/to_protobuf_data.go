package p2p

import (
	flow "github.com/dappledger/AnnChain/gemmill/modules/go-flowrate/flowrate"
	pbp2p "github.com/dappledger/AnnChain/gemmill/protos/p2p"
)

func (s *ChannelStatus) ToPbData() *pbp2p.ChannelStatus {
	if s == nil {
		return nil
	}
	pbStatus := &pbp2p.ChannelStatus{
		ID:                uint32(s.ID),
		SendQueueCapacity: int32(s.SendQueueCapacity),
		SendQueueSize:     int32(s.SendQueueSize),
		Priority:          int32(s.Priority),
		RecentlySent:      s.RecentlySent,
	}
	return pbStatus
}

func FlowStatusToPbData(s *flow.Status) *pbp2p.FlowStatus {
	if s == nil {
		return nil
	}
	pbStatus := &pbp2p.FlowStatus{
		Active:   s.Active,
		Start:    s.Start.UnixNano() / 1e6,
		Duration: s.Duration.Nanoseconds(),
		Idle:     s.Idle.Nanoseconds(),
		Bytes:    s.Bytes,
		Samples:  s.Samples,
		InstRate: s.InstRate,
		CurRate:  s.CurRate,
		AvgRate:  s.AvgRate,
		PeakRate: s.PeakRate,
		BytesRem: s.BytesRem,
		TimeRem:  s.TimeRem.Nanoseconds(),
		Progress: uint32(s.Progress),
	}
	return pbStatus
}

func (s *ConnectionStatus) ToPbData() *pbp2p.ConnectionStatus {
	if s == nil {
		return nil
	}
	pbStatus := &pbp2p.ConnectionStatus{
		SendMonitor: FlowStatusToPbData(&s.SendMonitor),
		RecvMonitor: FlowStatusToPbData(&s.RecvMonitor),
	}
	for _, v := range s.Channels {
		pbStatus.Channels = append(pbStatus.Channels, v.ToPbData())
	}
	return pbStatus
}

func (n *NodeInfo) ToPbData() (p *pbp2p.NodeInfo) {
	if n == nil {
		return nil
	}
	p = &pbp2p.NodeInfo{
		PubKey:      n.PubKey.ToPbData(),
		SigndPubKey: n.SigndPubKey,
		Moniker:     n.Moniker,
		Network:     n.Network,
		RemoteAddr:  n.RemoteAddr,
		ListenAddr:  n.ListenAddr,
		Version:     n.Version,
		Other:       n.Other,
	}
	return
}
