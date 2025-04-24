package clientscripts

import (
	"github.com/asek-ll/aecc-server/internal/dao"
)

type ScriptsManager struct {
	daos *dao.DaoProvider
}

func NewScriptsManager(daos *dao.DaoProvider) *ScriptsManager {
	return &ScriptsManager{
		daos: daos,
	}
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
