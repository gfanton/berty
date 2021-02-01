package proximitytransport

import (
	"fmt"

	"go.uber.org/zap"
)

// ReceiveFromPeer is called by native driver when peer's device sent data.
func (t *ProximityTransport) ReceiveFromPeer(remotePID string, payload []byte) {
	t.logger.Debug("ReceiveFromPeer()", zap.String("remotePID", remotePID))
	if err := t.pc.Pipe(payload, NewCustomAddr("p2p", remotePID)); err != nil {
		t.logger.Warn("receive from peer error", zap.Error(err), zap.String("remoteid", remotePID))
	}
}

func (t *ProximityTransport) WriteTo(p []byte, addr Addr) (n int, err error) {
	if !t.driver.SendToPeer(addr.String(), p) {
		t.logger.Debug("Conn.Write failed")
		// return n bytes
		return 0, fmt.Errorf("error: Conn.Write failed: native write failed")
	}

	return len(p), nil
}
