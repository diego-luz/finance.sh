import { LegalLayout, LegalSection } from '@/layouts/LegalLayout';

export function PrivacyPolicyPage() {
  return (
    <LegalLayout title="Política de Privacidade" updatedAt="26 de maio de 2026">
      <p>
        Esta Política de Privacidade descreve como o finance.sh ("nós") coleta, utiliza, armazena e
        protege dados pessoais, em conformidade com a Lei Geral de Proteção de Dados Pessoais
        (Lei nº 13.709/2018 - LGPD). Ao utilizar a Plataforma, você concorda com as práticas aqui
        descritas.
      </p>

      <LegalSection title="1. Dados que coletamos">
        <ul className="ml-5 list-disc space-y-1">
          <li>
            <strong>Dados de cadastro:</strong> nome, e-mail, senha (armazenada de forma
            criptografada) e o nome da organização.
          </li>
          <li>
            <strong>Dados financeiros inseridos:</strong> contas, transações, categorias,
            cartões, orçamentos e metas registrados por você.
          </li>
          <li>
            <strong>Dados de uso e segurança:</strong> registros de acesso, endereço IP,
            dispositivo e eventos de autenticação (incluindo 2FA).
          </li>
        </ul>
      </LegalSection>

      <LegalSection title="2. Finalidade do tratamento">
        <p>Utilizamos os dados para:</p>
        <ul className="ml-5 list-disc space-y-1">
          <li>fornecer, operar e manter a Plataforma e suas funcionalidades;</li>
          <li>autenticar usuários e proteger as contas contra acessos não autorizados;</li>
          <li>gerar relatórios e visualizações financeiras solicitadas por você;</li>
          <li>cumprir obrigações legais e regulatórias.</li>
        </ul>
      </LegalSection>

      <LegalSection title="3. Base legal">
        <p>
          O tratamento se fundamenta na <strong>execução de contrato</strong> (prestação do
          serviço), no <strong>consentimento</strong> fornecido no cadastro e para cookies, no{' '}
          <strong>cumprimento de obrigação legal</strong> e no <strong>legítimo interesse</strong>{' '}
          para fins de segurança e prevenção a fraudes.
        </p>
      </LegalSection>

      <LegalSection title="4. Cookies">
        <p>
          Utilizamos cookies estritamente necessários para autenticação e funcionamento do
          serviço, e cookies opcionais para preferências e melhoria de experiência. Você pode
          gerenciar sua escolha pelo banner de consentimento exibido no primeiro acesso.
        </p>
      </LegalSection>

      <LegalSection title="5. Compartilhamento">
        <p>
          Não vendemos dados pessoais. O compartilhamento ocorre apenas com operadores
          (fornecedores de infraestrutura e processamento) estritamente necessários à operação,
          sob obrigações contratuais de confidencialidade e segurança, ou quando exigido por lei.
        </p>
      </LegalSection>

      <LegalSection title="6. Retenção e eliminação">
        <p>
          Os dados são mantidos enquanto a conta estiver ativa ou pelo período necessário ao
          cumprimento das finalidades e de obrigações legais. Ao solicitar a exclusão da conta, os
          dados pessoais são eliminados ou anonimizados, ressalvadas as hipóteses de guarda
          obrigatória previstas em lei.
        </p>
      </LegalSection>

      <LegalSection title="7. Direitos do titular">
        <p>Nos termos da LGPD, você pode, a qualquer momento:</p>
        <ul className="ml-5 list-disc space-y-1">
          <li>confirmar a existência de tratamento e acessar seus dados;</li>
          <li>solicitar a correção de dados incompletos ou desatualizados;</li>
          <li>
            solicitar a portabilidade — disponível na Plataforma em{' '}
            <em>Configurações &gt; Privacidade e Dados &gt; Exportar meus dados</em>;
          </li>
          <li>
            solicitar a eliminação dos dados e o encerramento da conta, também disponível em{' '}
            <em>Configurações &gt; Privacidade e Dados</em>;
          </li>
          <li>revogar o consentimento e opor-se a tratamentos.</li>
        </ul>
      </LegalSection>

      <LegalSection title="8. Segurança">
        <p>
          Adotamos medidas técnicas e organizacionais para proteger os dados, incluindo
          criptografia de senhas, comunicação via HTTPS, controle de acesso por papéis e suporte
          a autenticação em dois fatores. Nenhum sistema é totalmente imune a riscos, mas
          trabalhamos continuamente para mitigá-los.
        </p>
      </LegalSection>

      <LegalSection title="9. Encarregado (DPO) e contato">
        <p>
          Para exercer seus direitos ou esclarecer dúvidas sobre o tratamento de dados, entre em
          contato com o nosso Encarregado pela Proteção de Dados (DPO):
        </p>
        <p>
          <strong>E-mail:</strong>{' '}
          <a href="mailto:dpo@finance.sh" className="text-primary hover:text-primary-600">
            dpo@finance.sh
          </a>
        </p>
      </LegalSection>

      <LegalSection title="10. Alterações desta política">
        <p>
          Esta Política pode ser atualizada para refletir mudanças legais ou no serviço. A versão
          vigente estará sempre disponível nesta página, com a data da última atualização.
        </p>
      </LegalSection>
    </LegalLayout>
  );
}
