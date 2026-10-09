import os
import subprocess

def run(cmd):
    subprocess.run(cmd, shell=True, check=True)

commits = [
    {
        "file": "internal/model/Workspace.go",
        "content": "package model\n\ntype Workspace struct {\n}\n",
        "msg": "feat(model): create Workspace model"
    },
    {
        "file": "internal/model/Workspace.go",
        "content": "package model\n\ntype Workspace struct {\n\tID string `json:\"id\"`\n}\n",
        "msg": "feat(model): add ID to Workspace"
    },
    {
        "file": "internal/model/Workspace.go",
        "content": "package model\n\ntype Workspace struct {\n\tID string `json:\"id\"`\n\tName string `json:\"name\"`\n}\n",
        "msg": "feat(model): add Name to Workspace"
    },
    {
        "file": "internal/model/Workspace.go",
        "content": "package model\n\nimport \"time\"\n\ntype Workspace struct {\n\tID string `json:\"id\"`\n\tName string `json:\"name\"`\n\tCreatedAt time.Time `json:\"createdAt\"`\n}\n",
        "msg": "feat(model): add CreatedAt to Workspace"
    },
    {
        "file": "internal/model/Workspace.go",
        "content": "package model\n\nimport \"time\"\n\ntype Workspace struct {\n\tID string `json:\"id\"`\n\tName string `json:\"name\"`\n\tCreatedAt time.Time `json:\"createdAt\"`\n\tUpdatedAt time.Time `json:\"updatedAt\"`\n}\n",
        "msg": "feat(model): add UpdatedAt to Workspace"
    },
    {
        "file": "internal/model/workspace_test.go",
        "content": "package model\n\nimport \"testing\"\n\nfunc TestWorkspace(t *testing.T) {\n}\n",
        "msg": "test(model): init Workspace tests"
    },
    {
        "file": "internal/model/workspace_test.go",
        "content": "package model\n\nimport \"testing\"\n\nfunc TestWorkspace(t *testing.T) {\n\tw := Workspace{Name: \"My Workspace\"}\n\tif w.Name != \"My Workspace\" {\n\t\tt.Errorf(\"expected My Workspace\")\n\t}\n}\n",
        "msg": "test(model): test Workspace basic fields"
    },
    {
        "file": "internal/service/workspace.go",
        "content": "package service\n\ntype WorkspaceService interface {\n}\n",
        "msg": "feat(service): init WorkspaceService interface"
    },
    {
        "file": "internal/service/workspace.go",
        "content": "package service\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/model\"\n\ntype WorkspaceService interface {\n\tCreate(w *model.Workspace) error\n}\n",
        "msg": "feat(service): add Create to WorkspaceService"
    },
    {
        "file": "internal/service/workspace.go",
        "content": "package service\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/model\"\n\ntype WorkspaceService interface {\n\tCreate(w *model.Workspace) error\n\tGet(id string) (*model.Workspace, error)\n}\n",
        "msg": "feat(service): add Get to WorkspaceService"
    },
    {
        "file": "internal/service/workspace.go",
        "content": "package service\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/model\"\n\ntype WorkspaceService interface {\n\tCreate(w *model.Workspace) error\n\tGet(id string) (*model.Workspace, error)\n\tList() ([]*model.Workspace, error)\n}\n",
        "msg": "feat(service): add List to WorkspaceService"
    },
    {
        "file": "internal/service/workspace.go",
        "content": "package service\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/model\"\n\ntype WorkspaceService interface {\n\tCreate(w *model.Workspace) error\n\tGet(id string) (*model.Workspace, error)\n\tList() ([]*model.Workspace, error)\n}\n\ntype workspaceServiceImpl struct {}\n",
        "msg": "feat(service): add workspaceServiceImpl struct"
    },
    {
        "file": "internal/service/workspace.go",
        "content": "package service\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/model\"\n\ntype WorkspaceService interface {\n\tCreate(w *model.Workspace) error\n\tGet(id string) (*model.Workspace, error)\n\tList() ([]*model.Workspace, error)\n}\n\ntype workspaceServiceImpl struct {}\n\nfunc NewWorkspaceService() WorkspaceService {\n\treturn &workspaceServiceImpl{}\n}\n",
        "msg": "feat(service): add NewWorkspaceService constructor"
    },
    {
        "file": "internal/service/workspace.go",
        "content": "package service\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/model\"\n\ntype WorkspaceService interface {\n\tCreate(w *model.Workspace) error\n\tGet(id string) (*model.Workspace, error)\n\tList() ([]*model.Workspace, error)\n}\n\ntype workspaceServiceImpl struct {}\n\nfunc NewWorkspaceService() WorkspaceService {\n\treturn &workspaceServiceImpl{}\n}\n\nfunc (s *workspaceServiceImpl) Create(w *model.Workspace) error { return nil }\n",
        "msg": "feat(service): stub WorkspaceService Create method"
    },
    {
        "file": "internal/service/workspace.go",
        "content": "package service\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/model\"\n\ntype WorkspaceService interface {\n\tCreate(w *model.Workspace) error\n\tGet(id string) (*model.Workspace, error)\n\tList() ([]*model.Workspace, error)\n}\n\ntype workspaceServiceImpl struct {}\n\nfunc NewWorkspaceService() WorkspaceService {\n\treturn &workspaceServiceImpl{}\n}\n\nfunc (s *workspaceServiceImpl) Create(w *model.Workspace) error { return nil }\nfunc (s *workspaceServiceImpl) Get(id string) (*model.Workspace, error) { return nil, nil }\n",
        "msg": "feat(service): stub WorkspaceService Get method"
    },
    {
        "file": "internal/service/workspace.go",
        "content": "package service\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/model\"\n\ntype WorkspaceService interface {\n\tCreate(w *model.Workspace) error\n\tGet(id string) (*model.Workspace, error)\n\tList() ([]*model.Workspace, error)\n}\n\ntype workspaceServiceImpl struct {}\n\nfunc NewWorkspaceService() WorkspaceService {\n\treturn &workspaceServiceImpl{}\n}\n\nfunc (s *workspaceServiceImpl) Create(w *model.Workspace) error { return nil }\nfunc (s *workspaceServiceImpl) Get(id string) (*model.Workspace, error) { return nil, nil }\nfunc (s *workspaceServiceImpl) List() ([]*model.Workspace, error) { return nil, nil }\n",
        "msg": "feat(service): stub WorkspaceService List method"
    },
    {
        "file": "internal/handler/workspace.go",
        "content": "package handler\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/service\"\n\ntype WorkspaceHandler struct {\n\tservice service.WorkspaceService\n}\n",
        "msg": "feat(handler): add WorkspaceHandler struct"
    },
    {
        "file": "internal/handler/workspace.go",
        "content": "package handler\n\nimport \"github.com/JohnMalugu/tsk-mgr-api/internal/service\"\n\ntype WorkspaceHandler struct {\n\tservice service.WorkspaceService\n}\n\nfunc NewWorkspaceHandler(s service.WorkspaceService) *WorkspaceHandler {\n\treturn &WorkspaceHandler{service: s}\n}\n",
        "msg": "feat(handler): add NewWorkspaceHandler constructor"
    },
    {
        "file": "internal/handler/workspace.go",
        "content": "package handler\n\nimport (\n\t\"net/http\"\n\t\"github.com/JohnMalugu/tsk-mgr-api/internal/service\"\n)\n\ntype WorkspaceHandler struct {\n\tservice service.WorkspaceService\n}\n\nfunc NewWorkspaceHandler(s service.WorkspaceService) *WorkspaceHandler {\n\treturn &WorkspaceHandler{service: s}\n}\n\nfunc (h *WorkspaceHandler) HandleList(w http.ResponseWriter, r *http.Request) {\n\tw.WriteHeader(http.StatusOK)\n}\n",
        "msg": "feat(handler): add HandleList stub for workspaces"
    },
    {
        "file": "internal/handler/workspace.go",
        "content": "package handler\n\nimport (\n\t\"net/http\"\n\t\"github.com/JohnMalugu/tsk-mgr-api/internal/service\"\n)\n\ntype WorkspaceHandler struct {\n\tservice service.WorkspaceService\n}\n\nfunc NewWorkspaceHandler(s service.WorkspaceService) *WorkspaceHandler {\n\treturn &WorkspaceHandler{service: s}\n}\n\nfunc (h *WorkspaceHandler) HandleList(w http.ResponseWriter, r *http.Request) {\n\tw.WriteHeader(http.StatusOK)\n}\n\nfunc (h *WorkspaceHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {\n\tw.WriteHeader(http.StatusCreated)\n}\n",
        "msg": "feat(handler): add HandleCreate stub for workspaces"
    }
]

for i, commit in enumerate(commits):
    os.makedirs(os.path.dirname(commit["file"]), exist_ok=True)
    with open(commit["file"], "w") as f:
        f.write(commit["content"])
    run(f"git add {commit['file']}")
    run(f"git commit -m \"{commit['msg']}\"")
    print(f"Committed {i+1}/20: {commit['msg']}")

run("go build ./...")
run("go test ./...")
