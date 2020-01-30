package irmaclient

import (
	"testing"

	irma "github.com/privacybydesign/irmago"
	"github.com/privacybydesign/irmago/internal/test"
	"github.com/stretchr/testify/require"
)

// Test pinchange interaction
func TestKeyshareChangePin(t *testing.T) {
	test.StartKeyshareServer(t)
	defer test.StopKeyshareServer(t)
	client, handler := parseStorage(t)
	defer test.ClearTestStorage(t handler.storage)

	require.NoError(t, client.keyshareChangePinWorker(irma.NewSchemeManagerIdentifier("test"), "12345", "54321"))
	require.NoError(t, client.keyshareChangePinWorker(irma.NewSchemeManagerIdentifier("test"), "54321", "12345"))
}
