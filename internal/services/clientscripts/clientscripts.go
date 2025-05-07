package clientscripts

import (
	"github.com/asek-ll/aecc-server/internal/dao"
	// "github.com/asek-ll/aecc-server/internal/wsmethods"
)

type OnUpdate func(role string, content string) error

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

func (m *ScriptsManager) CreateScript(role string) error {
	err := m.daos.ClientsScripts.CreateClientScript(&dao.ClientsScript{
		Role:    role,
		Content: "",
	})
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
		err = m.onUpdate(newRole, content)
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
