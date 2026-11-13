# TODO Consolidado (Pendências de `analyse.md` + `improve.md`)

Atualizado em: 2026-03-13

Legenda de prioridade:
- `P0` Crítico (segurança, risco de disponibilidade ou exposição indevida)
- `P1` Alto (impacto direto em performance, robustez ou operação)
- `P2` Médio (correção de comportamento e débito técnico relevante)
- `P3` Baixo (melhorias estruturais/refatoração)

| ID | Prioridade | Categoria | Tarefa | Arquivos-alvo | Origem | Status |
|---|---|---|---|---|---|---|
| ACL-REDIS-05 | P0 | Segurança/Autorização | Implementar invalidação explícita de cache ACL em: `AddMember`, `RemoveMember`, `AcceptInvitation`, alteração de `access_level` em grupo e `SoftDeleteGroup` (preferir estratégia por lista de membros). | `internal/acl/service.go`, `internal/acl/repository.go`, `adapter/acl/*` | improve | Não iniciado |
| JWKS-01 | P0 | Segurança/Resiliência | Limitar tamanho do body no fetch de JWKS com `io.LimitReader` para mitigar resposta excessiva/maliciosa. | `adapter/jwks/http_jwks.go` | analyse | Não iniciado |
| ENVLOADER-01 | P0 | Robustez | Eliminar risco de panic por type assertion nos getters do envloader (retornos seguros + erro/default). | `pkg/envloader/*` | analyse | Não iniciado |
| ACL-REDIS-01 | P1 | Performance | Adicionar dependência Redis (`github.com/redis/go-redis/v9`) e cliente/base de conexão. | `go.mod`, `go.sum`, `adapter/*` | improve | Não iniciado |
| ACL-REDIS-02 | P1 | Configuração | Adicionar `REDIS_URL` e configuração associada no carregamento de env da API. | `cmd/api/env/env.go`, `.env.example` | improve | Não iniciado |
| ACL-REDIS-03 | P1 | Performance/Autorização | Implementar `CachedAuthorizer` (cache-aside) para `CanAccess`, `GetEffectiveAccessLevel` e `ListAccessibleResourceIDs` usando chaves `acl:access:*` e `acl:resources:*` com TTLs definidos. | `adapter/acl/cache.go` (novo), `pkg/acl/*`, `internal/acl/*` | improve | Não iniciado |
| ACL-REDIS-04 | P1 | Segurança/Performance | Cachear sentinela `"none"` (sem acesso) para evitar cache-miss attack; considerar TTL menor para negativos. | `adapter/acl/cache.go` | improve | Não iniciado |
| ACL-REDIS-06 | P1 | Disponibilidade | Implementar fallback para banco quando Redis falhar (não bloquear autorização por indisponibilidade de cache). | `adapter/acl/cache.go` | improve | Não iniciado |
| ACL-REDIS-07 | P1 | Arquitetura/Wiring | Fazer composição do authorizer com cache na camada de fachada/wiring (`cmd/api/*`), mantendo `main.go` declarativo. | `cmd/api/acl/route.go`, `cmd/api/bootstrap.go` (ou factory dedicado) | improve | Não iniciado |
| LIMITER-01 | P1 | Escalabilidade | Implementar limiter distribuído com Redis (estado compartilhado entre instâncias). | `cmd/api/mid/limiter.go`, `adapter/*` | analyse | Não iniciado |
| ROUTER-ERR-01 | P1 | API/Observabilidade | Adicionar `ErrorHandler` custom no config do Fiber para padronizar erros de framework e serialização de resposta. | `pkg/router/router.go` | analyse | Não iniciado |
| AUTH-01 | P2 | Compatibilidade | Tornar validação do prefixo `Bearer` case-insensitive em autenticação. | `pkg/auth/auth.go` | analyse | Não iniciado |
| NOTFOUND-01 | P2 | Organização | Remover middleware `NotFound` órfão ou integrá-lo corretamente na cadeia de rotas. | `cmd/api/mid/notfound.go`, `cmd/api/main.go` | analyse | Não iniciado |
| HELLO-01 | P2 | Manutenibilidade | Remover duplicação de criação de `NewService` nas rotas de `hello` (centralizar construção). | `internal/hello/*`, `cmd/api/main.go` | analyse | Não iniciado |
| TX-01 | P2 | Confiabilidade | Revisar rollback redundante após `Commit` com erro e simplificar fluxo transacional. | `internal/*`, `adapter/db/*` | analyse | Não iniciado |
| MID-WRAPPER-01 | P3 | Arquitetura | Revisar wrappers em `cmd/api/mid` que apenas delegam; manter somente os que agregam valor real de fachada/customização. | `cmd/api/mid/*` | analyse | Não iniciado |

