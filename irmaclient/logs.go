package irmaclient

import (
	"encoding/json"
	"time"

	"github.com/bwesterb/go-atum"
	"github.com/go-errors/errors"
	"github.com/privacybydesign/irmago"
)

// LogEntry is a log entry of a past event.
type LogEntry struct {
	// General info
	Type    irma.Action
	Time    irma.Timestamp        // Time at which the session was completed
	Version *irma.ProtocolVersion `json:",omitempty"` // Protocol version that was used in the session

	Request json.RawMessage     `json:",omitempty"` // Message that started the session
	request irma.SessionRequest // cached parsed version of Request; get with LogEntry.SessionRequest()

	// Session type-specific info
	Removed                map[irma.CredentialTypeIdentifier][]irma.TranslatedString `json:",omitempty"` // In case of credential removal
	SignedMessage          []byte                                                    `json:",omitempty"` // In case of signature sessions
	Timestamp              *atum.Timestamp                                           `json:",omitempty"` // In case of signature sessions
	SignedMessageLDContext string                                                    `json:",omitempty"` // In case of signature sessions

	IssueCommitment *irma.IssueCommitmentMessage `json:",omitempty"`
	Disclosure      *irma.Disclosure             `json:",omitempty"`
}

const actionRemoval = irma.Action("removal")

func (entry *LogEntry) SessionRequest() (irma.SessionRequest, error) {
	if entry.request == nil {
		switch entry.Type {
		case irma.ActionDisclosing:
			entry.request = &irma.DisclosureRequest{}
		case irma.ActionSigning:
			entry.request = &irma.SignatureRequest{}
		case irma.ActionIssuing:
			entry.request = &irma.IssuanceRequest{}
		default:
			return nil, nil
		}
	}

	err := json.Unmarshal([]byte(entry.Request), entry.request)
	if err != nil {
		return nil, err
	}

	return entry.request, nil
}

func (entry *LogEntry) setSessionRequest() error {
	bts, err := json.Marshal(entry.request)
	if err != nil {
		return err
	}
	entry.Request = json.RawMessage(bts)
	return nil
}

// GetDisclosedCredentials gets the list of disclosed credentials for a log entry
func (entry *LogEntry) GetDisclosedCredentials(conf *irma.Configuration) ([]*irma.DisclosedAttribute, error) {
	if entry.Type == actionRemoval {
		return []*irma.DisclosedAttribute{}, nil
	}

	request, err := entry.SessionRequest()
	if err != nil {
		return nil, err
	}
	var disclosure *irma.Disclosure
	disjunctions := request.ToDisclose()
	if entry.Type == irma.ActionIssuing {
		disclosure = entry.IssueCommitment.Disclosure()
	} else {
		disclosure = entry.Disclosure
	}
	_, attrs, err := disclosure.DisclosedAttributes(conf, disjunctions)
	return attrs, err
}

// GetIssuedCredentials gets the list of issued credentials for a log entry
func (entry *LogEntry) GetIssuedCredentials(conf *irma.Configuration) (list irma.CredentialInfoList, err error) {
	if entry.Type != irma.ActionIssuing {
		return irma.CredentialInfoList{}, nil
	}
	request, err := entry.SessionRequest()
	if err != nil {
		return nil, err
	}
	return request.(*irma.IssuanceRequest).GetCredentialInfoList(conf, entry.Version)
}

// GetSignedMessage gets the signed for a log entry
func (entry *LogEntry) GetSignedMessage() (abs *irma.SignedMessage, err error) {
	if entry.Type != irma.ActionSigning {
		return nil, nil
	}
	request, err := entry.SessionRequest()
	if err != nil {
		return nil, err
	}
	sigrequest := request.(*irma.SignatureRequest)
	return &irma.SignedMessage{
		LDContext: entry.SignedMessageLDContext,
		Signature: entry.Disclosure.Proofs,
		Nonce:     sigrequest.Nonce,
		Context:   sigrequest.Context,
		Message:   string(entry.SignedMessage),
		Timestamp: entry.Timestamp,
	}, nil
}

func (session *session) createLogEntry(response interface{}) (*LogEntry, error) {
	entry := &LogEntry{
		Type:    session.Action,
		Time:    irma.Timestamp(time.Now()),
		Version: session.Version,
		request: session.request,
	}

	if err := entry.setSessionRequest(); err != nil {
		return nil, err
	}

	switch entry.Type {
	case actionRemoval:

	case irma.ActionSigning:
		// Get the signed message and timestamp
		request := session.request.(*irma.SignatureRequest)
		entry.SignedMessage = []byte(request.Message)
		entry.Timestamp = request.Timestamp
		entry.SignedMessageLDContext = irma.SignedMessageLDContext

		fallthrough
	case irma.ActionDisclosing:
		entry.Disclosure = response.(*irma.Disclosure)
	case irma.ActionIssuing:
		entry.IssueCommitment = response.(*irma.IssueCommitmentMessage)
	default:
		return nil, errors.New("Invalid log type")
	}

	return entry, nil
}
