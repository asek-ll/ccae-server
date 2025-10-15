package clientscripts

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	// "github.com/asek-ll/aecc-server/internal/wsmethods"
)

type OnUpdate func(*dao.ClientsScript) error

type ScriptJsonView struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Version int    `json:"version"`
}

func (sjv *ScriptJsonView) ToScript() *dao.ClientsScript {
	return &dao.ClientsScript{
		Role:    sjv.Role,
		Content: sjv.Content,
		Version: sjv.Version,
	}
}

func NewScriptJsonView(script *dao.ClientsScript) *ScriptJsonView {
	return &ScriptJsonView{
		Role:    script.Role,
		Content: script.Content,
		Version: script.Version,
	}
}

type ScriptsManager struct {
	daos     *dao.DaoProvider
	onUpdate OnUpdate
}

func NewScriptsManager(daos *dao.DaoProvider) *ScriptsManager {
	return &ScriptsManager{
		daos: daos,
	}
}

func (m *ScriptsManager) SetOnUpdate(f OnUpdate) {
	m.onUpdate = f
}

func (m *ScriptsManager) CreateScript(script *dao.ClientsScript) error {
	err := m.daos.ClientsScripts.CreateClientScript(script)
	return err
}

func (m *ScriptsManager) UpdateScript(role string, newRole string, content string) error {
	err := m.daos.ClientsScripts.UpdateClientsScript(role, &dao.ClientsScript{
		Role:    newRole,
		Content: content,
	})

	if err != nil {
		return err
	}

	if m.onUpdate != nil {
		script, err := m.daos.ClientsScripts.GetClientScript(role)
		if err != nil {
			return err
		}
		err = m.onUpdate(script)
		if err != nil {
			return err
		}
	}

	return err
}

func (m *ScriptsManager) DeleteScript(role string) error {
	err := m.daos.ClientsScripts.DeleteClientScript(role)
	return err
}

func (m *ScriptsManager) GetScript(role string) (*dao.ClientsScript, error) {
	script, err := m.daos.ClientsScripts.GetClientScript(role)
	return script, err
}

func (m *ScriptsManager) GetScripts() ([]*dao.ClientsScript, error) {
	scripts, err := m.daos.ClientsScripts.GetClientsScripts()
	return scripts, err
}
