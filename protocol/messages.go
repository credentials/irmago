package protocol

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/credentials/irmago"
)

// Status encodes the status of an IRMA session (e.g., connected).
type Status string

// Action encodes the session type of an IRMA session (e.g., disclosing).
type Action string

// Version encodes the IRMA protocol version of an IRMA session.
type Version string

// Statuses
const (
	StatusConnected     = Status("connected")
	StatusCommunicating = Status("communicating")
)

// Actions
const (
	ActionDisclosing = Action("disclosing")
	ActionSigning    = Action("signing")
	ActionIssuing    = Action("issuing")
	ActionUnknown    = Action("unknown")
)

// Qr contains the data of an IRMA session QR (as generated by irma_js),
// suitable for NewSession().
type Qr struct {
	// Server with which to perform the session
	URL string `json:"u"`
	// Session type (disclosing, signing, issuing)
	Type               Action `json:"irmaqr"`
	ProtocolVersion    string `json:"v"`
	ProtocolMaxVersion string `json:"vmax"`
}

// A SessionInfo is the first message in the IRMA protocol (i.e., GET on the server URL),
// containing the session request info.
type SessionInfo struct {
	Jwt     string                          `json:"jwt"`
	Nonce   *big.Int                        `json:"nonce"`
	Context *big.Int                        `json:"context"`
	Keys    map[irmago.IssuerIdentifier]int `json:"keys"`
}

/*
So apparently, in the old Java implementation we forgot to write a (de)serialization for the Java
equivalent of the type IssuerIdentifier. This means a Java IssuerIdentifier does not serialize to
a string, but to e.g. `{"identifier":"irma-demo.RU"}`.
This is a complex data type, so not suitable to act as keys in a JSON map. Consequentially,
Gson serializes the `json:"keys"` field not as a map, but as a list consisting of pairs where
the first item of the pair is a serialized IssuerIdentifier as above, and the second item
of the pair is the corresponding key counter from the original map.
This is a bit of a mess to have to deserialize. See below. In a future version of the protocol,
this will have to be fixed both in the Java world and here in Go.
*/

type jsonSessionInfo struct {
	Jwt     string          `json:"jwt"`
	Nonce   *big.Int        `json:"nonce"`
	Context *big.Int        `json:"context"`
	Keys    [][]interface{} `json:"keys"`
}

// UnmarshalJSON unmarshals session information.
func (si *SessionInfo) UnmarshalJSON(b []byte) error {
	temp := &jsonSessionInfo{}
	err := json.Unmarshal(b, temp)
	if err != nil {
		return err
	}

	si.Jwt = temp.Jwt
	si.Nonce = temp.Nonce
	si.Context = temp.Context
	si.Keys = make(map[irmago.IssuerIdentifier]int, len(temp.Keys))
	for _, item := range temp.Keys {
		var idmap map[string]interface{}
		var idstr string
		var counter float64
		var ok bool
		if idmap, ok = item[0].(map[string]interface{}); !ok {
			return errors.New("Failed to deserialize session info")
		}
		if idstr, ok = idmap["identifier"].(string); !ok {
			return errors.New("Failed to deserialize session info")
		}
		if counter, ok = item[1].(float64); !ok {
			return errors.New("Failed to deserialize session info")
		}
		id := irmago.NewIssuerIdentifier(idstr)
		si.Keys[id] = int(counter)
	}
	return nil
}
