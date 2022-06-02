package rest

import "strings"

type AdRequest struct {
	// List of important key words from the current page describing the context.
	ContextKeys []string `json:"context_keys,omitempty"`
}

func (m *AdRequest) GetContextkeys() []string {
	return m.ContextKeys
}

func (m *AdRequest) GetContextKeysList() string {
	return strings.Join(m.ContextKeys, ",")
}

func (m *AdRequest) GetContextKeysCount() int {
	return len(m.ContextKeys)
}

type AdResponse struct {
	Ads []*Ad `json:"ads,omitempty"`
}

func (m *AdResponse) GetAds() []*Ad {
	if m != nil {
		return m.Ads
	}
	return nil
}

type Ad struct {
	// url to redirect to when an ad is clicked.
	RedirectUrl string `json:"redirect_url,omitempty"`
	// short advertisement text to display.
	Text string `json:"text,omitempty"`
}
