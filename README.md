# CRUD com Go e SQLite

Projeto completo de um sistema de gerenciamento de tarefas (CRUD) usando Go, SQLite e net/http, com interface web simples.

## Funcionalidades

- Criar, listar, buscar, atualizar e deletar tarefas via API REST.
- Interface web amigável (HTML/CSS/JS) que consome a API.
- Validações de entrada (título obrigatório, ID válido).
- Logging de todas as requisições HTTP.
- Timeouts configurados no servidor (leitura, escrita, idle).
- Porta configurável via variável de ambiente PORT.

## Como rodar

1. Clone ou baixe os arquivos main.go e index.html para a mesma pasta.
2. No terminal, execute:

go mod tidy
go run .

O servidor iniciará na porta 8080 (ou na porta definida em PORT).

## Acessando

- API REST: http://localhost:8080/tasks
- Interface web: http://localhost:8080/

## Rotas da API

| Método | Rota           | Descrição               |
|--------|----------------|-------------------------|
| GET    | /tasks         | Lista todas as tarefas  |
| POST   | /tasks         | Cria uma nova tarefa    |
| GET    | /tasks/{id}    | Busca uma tarefa por ID |
| PUT    | /tasks/{id}    | Atualiza uma tarefa     |
| DELETE | /tasks/{id}    | Apaga uma tarefa        |

## Exemplos com curl

### Criar tarefa
curl -X POST http://localhost:8080/tasks -H "Content-Type: application/json" -d '{"title":"Estudar Go","done":false}'

### Listar tarefas
curl http://localhost:8080/tasks

### Buscar por ID
curl http://localhost:8080/tasks/1

### Atualizar tarefa
curl -X PUT http://localhost:8080/tasks/1 -H "Content-Type: application/json" -d '{"title":"Estudar Go e SQLite","done":true}'

### Excluir tarefa
curl -X DELETE http://localhost:8080/tasks/1

## Variáveis de ambiente

- PORT: define a porta do servidor (padrão: 8080). Exemplo:
  PORT=3000 go run .

## Estrutura dos arquivos

.
├── main.go          # Servidor Go com rotas, banco SQLite e lógica CRUD
├── index.html       # Frontend simples para gerenciar tarefas
├── tasks.db         # Banco de dados SQLite (criado automaticamente)
└── README.md

## Melhorias implementadas

- Middleware de logging (mostra método e caminho de cada requisição)
- Timeouts: ReadTimeout, WriteTimeout, IdleTimeout
- Porta via variável de ambiente
- Validação de título vazio e ID inválido
- Tratamento adequado de erros HTTP (400, 404, 500)

## Tecnologias utilizadas

- Go (pacotes net/http, database/sql, encoding/json)
- SQLite com driver modernc.org/sqlite
- HTML/CSS/JS puro (frontend)

## Próximos passos (sugestões)

- Adicionar paginação e filtros na listagem
- Suporte a CORS para chamadas de outros domínios
- Testes automatizados
- Dockerizar a aplicação

## Licença

Este projeto é livre para estudo e uso.