# LGPD — Conformidade (finance.sh)

Mapeamento da **Lei Geral de Proteção de Dados** (Lei nº 13.709/2018) aos
controles do finance.sh (open-source self-hosted, AGPL-3.0).
Documento vivo; itens pendentes ao final.

---

## Controlador dos dados

Em finance.sh (self-hosted), o **controlador** dos dados é o
**operador** — a pessoa, MEI ou organização que instala e roda esta aplicação
em sua infraestrutura própria. Os mantenedores do projeto finance.sh são
apenas autores do software, **não** controladores nem operadores dos dados
armazenados em instâncias de terceiros.

Implicações:

- **Você** (operador) é responsável por DPO/RIPD/inventário (ROPA) da sua instância.
- **Você** atende aos direitos do titular (LGPD art. 18) usando as ferramentas
  deste software (exportar, excluir, audit log, consentimento versionado).
- **Você** mantém os backups, controla acessos, comunica incidentes à ANPD se
  aplicável.

Este documento, portanto, é um **guia** para o operador montar a sua própria
conformidade — não uma declaração dos mantenedores. As tabelas abaixo descrevem
o que o software **expõe** ao operador para que ele preencha o seu ROPA, defina
o seu DPO e responda à ANPD.

---

## 1. Bases legais (art. 7º / 11)

| Tratamento | Base legal |
|---|---|
| Cadastro e autenticação do usuário | Execução de contrato (art. 7º, V) |
| Operação do serviço financeiro (contas, transações, relatórios) | Execução de contrato (art. 7º, V) |
| Envio de notificações transacionais ao próprio titular (alertas de saldo, vencimentos cadastrados pelo usuário) | Execução de contrato / legítimo interesse (art. 7º, V/IX) |
| Logs de auditoria e segurança | Cumprimento de obrigação legal / legítimo interesse (art. 7º, II/IX) |
| Comunicações de marketing (se houver) | Consentimento (art. 7º, I) — opt-in separado |

Não há tratamento de dados sensíveis financeiros sob a ótica do art. 11 por
padrão; caso surjam (ex.: dados de saúde em notas), exigir consentimento
específico.

---

## 2. Inventário de dados tratados (ROPA — Registro de Operações)

| Dado | Categoria | Finalidade | Retenção | Onde / proteção |
|---|---|---|---|---|
| Nome, e-mail | Identificação/contato | Conta, login, contato | Enquanto durar a conta + `RETENTION_DAYS` após exclusão | Postgres; e-mail mascarado em logs |
| Senha | Credencial | Autenticação | Idem | Hash **bcrypt** (nunca em claro) |
| Segredo 2FA (TOTP) | Credencial | 2FA | Idem | **Cifrado AES-256-GCM** em campo |
| Token de refresh | Sessão | Sessão/segurança | Até expirar/rotacionar | **Hash** no banco |
| Organização, papel (RBAC) | Cadastro | Multi-organização/autorização | Idem conta | Postgres, escopado por `organization_id` |
| Contas, transações, orçamentos, metas | Financeiro | Funcionalidade central | Enquanto durar a conta + `RETENTION_DAYS` | Postgres, escopado por `organization_id` |
| Notas/observações (podem conter PII) | PII livre | Anotações do usuário | Idem | **Cifrado em campo** quando sensível |
| Logs de auditoria | Registro | Segurança/obrigação legal | Conforme política (ex.: 6 meses) | Postgres; PII mascarada |
| Versão de consentimento (`TERMS_VERSION`) | Consentimento | Prova de aceite | Permanente (prova) | Postgres |
| IP / metadados de acesso | Técnico | Segurança/rate limit | Curto prazo | Logs com rotação |

---

## 3. Direitos do titular (art. 18) e como o sistema atende

| Direito | Como o finance.sh atende |
|---|---|
| **Confirmação e acesso** | Endpoint de perfil (`/me`) e exportação. |
| **Portabilidade** | `GET /me/export` — exporta os dados do titular em formato estruturado (JSON). |
| **Eliminação / anonimização** | `DELETE /me/account` — remove/anonimiza a conta e dados associados; soft delete seguido de **purga** após `RETENTION_DAYS`. |
| **Correção** | Edição de dados de perfil e registros via API/UI. |
| **Revogação de consentimento** | Gerenciamento de opt-ins; consentimento **versionado** (`TERMS_VERSION`) para rastrear aceite/revogação. |
| **Informação sobre compartilhamento** | Documentado nesta política; operadores/subprocessadores listados quando aplicável. |
| **Revisão de decisões automatizadas** | Não há decisão automatizada com efeito jurídico relevante; havendo, prever revisão humana. |

---

## 4. Princípios aplicados

- **Minimização** (art. 6º, III): coletamos apenas o necessário (nome, e-mail,
  dados financeiros que o próprio usuário insere).
- **Finalidade / necessidade**: cada dado tem finalidade declarada no ROPA acima.
- **Retenção limitada**: o **scheduler in-process** purga dados expirados/soft-
  deletados após `RETENTION_DAYS` (configurável).
- **Segurança** (art. 46): TLS, criptografia em campo + em disco, RBAC,
  isolamento por organização, lockout, auditoria — ver [`SECURITY.md`](SECURITY.md).
- **Mascaramento de PII em logs**: e-mails/identificadores são mascarados.
- **Consentimento versionado**: aceite registrado com `TERMS_VERSION`.

---

## 5. Resposta a incidentes (art. 48)

1. **Detecção/contenção** — isolar o componente afetado; preservar evidências
   (logs de auditoria, logs de container).
2. **Avaliação** — escopo, dados e titulares afetados, risco/dano.
3. **Comunicação à ANPD e aos titulares** — em **prazo razoável (referência: 72h)**
   quando houver risco/dano relevante, com: natureza dos dados, titulares
   afetados, medidas técnicas adotadas, riscos e medidas de mitigação.
4. **Remediação e lições aprendidas** — correção da causa raiz, atualização de
   controles e deste documento.

Responsável pelo processo: **Encarregado/DPO** (placeholder no topo).

---

## 6. Subprocessadores / transferências

**O operador** lista seus próprios subprocessadores
(hospedagem/VPS, e-mail/SMTP, eventuais integrações que ele mesmo configurar)
e o país de processamento. Garantir cláusulas contratuais e, em transferência
internacional, base do art. 33. Os mantenedores do projeto **não** são
subprocessadores em deploys self-hosted.

---

## 7. Itens pendentes

Pendências divididas entre **operador** (o que cada deploy precisa fazer) e
**projeto** (o que os mantenedores ainda devem entregar no software):

**Para o operador:**

- [ ] Definir e publicar **Encarregado/DPO** da instância (nome + canal de contato).
- [ ] Publicar **Política de Privacidade** e **Termos** versionados da sua
      instância (e fixar o número em `TERMS_VERSION`).
- [ ] Definir prazos de retenção por categoria (especialmente logs de auditoria).
- [ ] Registro formal de **subprocessadores** e DPAs.
- [ ] **RIPD/DPIA** (Relatório de Impacto) antes de novos tratamentos de risco.
- [ ] Procedimento operacional documentado para atender pedidos de titular
      dentro do prazo legal.

**Para o projeto (software):**

- [ ] Tela de **gestão de consentimento** e histórico de versões na UI.
