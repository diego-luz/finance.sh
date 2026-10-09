# Roadmap — finance.sh

Sugira features abrindo uma
[issue de feature request](https://github.com/diego-luz/finance.sh/issues/new?template=feature_request.yml).

## v0.1 — Foundation (atual)

- ✅ Auth: JWT com rotação e detecção de reuso, 2FA TOTP, bloqueio por tentativas, verificação de e-mail
- ✅ RBAC + multi-organização + audit log + soft delete
- ✅ Setup wizard no primeiro acesso, com código de instalação
- ✅ Contas, transações, categorias, contatos
- ✅ Cartões + faturas + parcelamento
- ✅ Contas a pagar/receber
- ✅ Orçamentos + metas
- ✅ Recorrência via scheduler in-process
- ✅ Multi-moeda (15 ISO-4217)
- ✅ Anexos (BYTEA no Postgres)
- ✅ Import OFX/CSV com dedup (inclusive arquivos em Windows-1252)
- ✅ Busca global + tags + categorização automática
- ✅ Dashboard + projeção de fluxo + relatórios (Excel/PDF/CSV)
- ✅ PWA + i18n pt-BR/en/es + modo privacidade + dark mode
- ✅ Criptografia AES-GCM + telas de LGPD
- ✅ Screenshots + GIF de demonstração
- ✅ CI publicando a imagem em `ghcr.io/diego-luz/finance-sh-app`
- ✅ Testes de unidade no frontend (Vitest)

## v0.2 — Polish (próximo)

- [ ] Página de aceite de convite no frontend
- [ ] Notificações por usuário (hoje são da organização inteira)
- [ ] Trilha de auditoria de eventos de segurança (logins, bloqueios, 2FA)
- [ ] Testes de integração com Postgres no CI
- [ ] Migração para React Router 7 e Tailwind 4
- [ ] SQLite como opção para instalações leves (`DB_DRIVER=sqlite|postgres`)
- [ ] Guias de deploy em `docs/` (Portainer, Coolify, Synology, Unraid)
- [ ] Helm chart para Kubernetes
- [ ] One-click deploy: Railway, Render, Fly.io

## v1.0 — Estável

- [ ] API estável e versionada
- [ ] Telemetria opt-in anônima (estilo Plausible/Matomo)
- [ ] Open Finance Brasil (Pluggy/Belvo)
- [ ] OCR de comprovantes
- [ ] App mobile nativo

## Backlog

- [ ] WebSockets/push para notificações em tempo real
- [ ] OAuth (Google, GitHub, login social)
- [ ] Filas dedicadas (RabbitMQ/NATS), se necessário
