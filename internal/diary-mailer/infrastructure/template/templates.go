package template

import (
	"bytes"
	"fmt"
	htmltmpl "html/template"
	"text/template"
)

// TemplateStore provides template retrieval
type TemplateStore interface {
	Get(id, locale string) (*TemplateWrapper, error)
}

// TemplateWrapper holds subject, text and html templates
type TemplateWrapper struct {
	ID       string
	Subject  string
	TextTmpl *template.Template
	HTMLTmpl *htmltmpl.Template
}

// Render returns text and html rendered bodies
func (t *TemplateWrapper) Render(payload map[string]interface{}) (text string, html string, err error) {
	var tb bytes.Buffer
	if t.TextTmpl != nil {
		if err := t.TextTmpl.Execute(&tb, payload); err != nil {
			return "", "", err
		}
	}

	var hb bytes.Buffer
	if t.HTMLTmpl != nil {
		if err := t.HTMLTmpl.Execute(&hb, payload); err != nil {
			return "", "", err
		}
	}

	return tb.String(), hb.String(), nil
}

// InMemoryStore is a simple template store with HTML + text support
type InMemoryStore struct {
	templates map[string]*TemplateWrapper // key: id:locale
}

func NewInMemoryStore() *InMemoryStore {
	s := &InMemoryStore{templates: make(map[string]*TemplateWrapper)}
	// family invite: invited person
	familyInviteText := template.Must(template.New("family_invite_text_jp").Parse("{{.inviter_name}}さんから招待されています。\n\nログインして、{{.family_name}}へ参加して家族で日記を共有しましょう。\nURL: {{.app_url}}"))
	familyInviteHTML := htmltmpl.Must(htmltmpl.New("family_invite_html_jp").Parse("<html><body><p>{{.inviter_name}}さんから招待されています。</p><p>以下のリンクからログインし、<strong>{{.family_name}}</strong>へ参加して家族で日記を共有しましょう。</p><p><a href=\"{{.app_url}}\">{{.app_url}}</a></p></body></html>"))
	s.templates["family_invite_v1:ja"] = &TemplateWrapper{
		ID:       "family_invite_v1",
		Subject:  "fam-diary-logで、{{.inviter_name}}から招待されています。",
		TextTmpl: familyInviteText,
		HTMLTmpl: familyInviteHTML,
	}

	// family join request: someone requested to join
	familyRequestText := template.Must(template.New("family_request_text_jp").Parse("{{.requester_name}}さんが参加を申請しています。\n\n管理画面で新しいメンバーの参加申請を承諾し、日記を共有しましょう。\nURL: {{.app_url}}"))
	familyRequestHTML := htmltmpl.Must(htmltmpl.New("family_request_html_jp").Parse("<html><body><p>{{.requester_name}}さんが参加を申請しています。</p><p>{{.app_url}}で、新しいメンバーの参加申請を承諾し、日記を共有しましょう。</p><p><a href=\"{{.app_url}}\">{{.app_url}}</a></p></body></html>"))
	s.templates["family_request_v1:ja"] = &TemplateWrapper{
		ID:       "family_request_v1",
		Subject:  "fam-diary-logで、{{.requester_name}}が参加を申請しています。",
		TextTmpl: familyRequestText,
		HTMLTmpl: familyRequestHTML,
	}

	return s
}

func (s *InMemoryStore) Get(id, locale string) (*TemplateWrapper, error) {
	key := fmt.Sprintf("%s:%s", id, locale)
	if t, ok := s.templates[key]; ok {
		return t, nil
	}
	// fallback to default locale
	key = fmt.Sprintf("%s:en", id)
	if t, ok := s.templates[key]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("template not found: %s/%s", id, locale)
}
