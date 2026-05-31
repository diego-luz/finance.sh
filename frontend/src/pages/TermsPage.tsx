import { LegalLayout, LegalSection } from '@/layouts/LegalLayout';

export function TermsPage() {
  return (
    <LegalLayout title="Termos de Uso" updatedAt="26 de maio de 2026">
      <p>
        Estes Termos de Uso regem o acesso e a utilização da plataforma finance.sh ("Plataforma"),
        uma ferramenta de gestão financeira multiempresa. Ao criar uma conta ou utilizar a
        Plataforma, você declara ter lido, compreendido e concordado com as condições abaixo.
      </p>

      <LegalSection title="1. Objeto">
        <p>
          O finance.sh oferece ferramentas para organização de contas, transações, orçamentos, metas
          e relatórios financeiros de organizações. O finance.sh Community Edition é disponibilizado
          como software open-source (AGPL-3.0) para auto-hospedagem (self-hosted). Você instala em
          sua própria infraestrutura e é o controlador dos seus dados.
        </p>
      </LegalSection>

      <LegalSection title="2. Cadastro e conta">
        <p>
          Para utilizar a Plataforma é necessário criar uma conta com informações verídicas,
          completas e atualizadas. Você é responsável por manter a confidencialidade das suas
          credenciais e por todas as atividades realizadas em sua conta. Recomendamos fortemente
          a ativação da autenticação em dois fatores (2FA).
        </p>
      </LegalSection>

      <LegalSection title="3. Organizações, papéis e operação self-hosted">
        <p>
          Cada organização pode ter membros com diferentes papéis (proprietário, administrador,
          membro e visualizador), com permissões distintas. O operador (você, ao auto-hospedar) é
          responsável pela gestão da instância, backup, segurança e conformidade legal local.
        </p>
      </LegalSection>

      <LegalSection title="4. Uso aceitável">
        <p>Você concorda em não utilizar a Plataforma para:</p>
        <ul className="ml-5 list-disc space-y-1">
          <li>violar leis aplicáveis ou direitos de terceiros;</li>
          <li>inserir dados fraudulentos, ilícitos ou de origem não autorizada;</li>
          <li>
            tentar acessar áreas restritas, comprometer a segurança ou interromper o
            funcionamento do serviço;
          </li>
          <li>realizar engenharia reversa ou revender o serviço sem autorização.</li>
        </ul>
      </LegalSection>

      <LegalSection title="5. Dados e privacidade">
        <p>
          O tratamento de dados pessoais é regido pela nossa Política de Privacidade, elaborada
          em conformidade com a Lei Geral de Proteção de Dados (Lei nº 13.709/2018 - LGPD). Os
          dados financeiros inseridos pertencem à respectiva organização e ao titular.
        </p>
      </LegalSection>

      <LegalSection title="6. Disponibilidade e suporte">
        <p>
          Empregamos esforços razoáveis para manter a Plataforma disponível, mas o serviço pode
          sofrer interrupções para manutenção, atualizações ou por fatores fora do nosso
          controle. Não garantimos disponibilidade ininterrupta.
        </p>
      </LegalSection>

      <LegalSection title="7. Limitação de responsabilidade">
        <p>
          A Plataforma é uma ferramenta de apoio à gestão financeira e não constitui
          aconselhamento contábil, jurídico ou de investimentos. Na máxima extensão permitida em
          lei, não nos responsabilizamos por decisões tomadas com base nas informações exibidas.
        </p>
      </LegalSection>

      <LegalSection title="8. Encerramento">
        <p>
          Você pode encerrar sua conta a qualquer momento pelas configurações da Plataforma. Em
          caso de violação destes Termos, podemos suspender ou encerrar o acesso, observados os
          direitos do titular previstos na LGPD.
        </p>
      </LegalSection>

      <LegalSection title="9. Alterações">
        <p>
          Estes Termos podem ser atualizados periodicamente. Alterações relevantes serão
          comunicadas pelos canais disponíveis. O uso continuado após a publicação representa a
          concordância com a versão vigente.
        </p>
      </LegalSection>

      <LegalSection title="10. Contato">
        <p>
          Dúvidas sobre estes Termos podem ser encaminhadas para{' '}
          <a href="mailto:contato@finance.sh" className="text-primary hover:text-primary-600">
            contato@finance.sh
          </a>
          .
        </p>
      </LegalSection>
    </LegalLayout>
  );
}
