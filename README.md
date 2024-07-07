# Terminal Chat Client em Go
Este é um cliente de chat em terminal desenvolvido em Go, que se conecta a um servidor de chat utilizando websockets para comunicação em tempo real.

## Sobre o Projeto
O `terminal_chat_client_golang` foi criado para oferecer uma interface simples e interativa para comunicação em salas de chat via terminal, utilizando websockets em Go. 

## Funcionalidades Principais
- **Conexão em Tempo Real:** Conecta-se a um servidor de chat em tempo real utilizando websockets.

- **Envio de Mensagens:** Permite que o usuário envie mensagens para a sala de chat ativa.

- **Recepção de Mensagens:** Exibe mensagens recebidas de outros usuários na mesma sala.

- **Configuração Flexível:** Permite configurar o endereço do servidor através de flags de linha de comando.

## Objetivos
- **Exploração de Websockets:** Implementação de interação bidirecional instantânea entre cliente e servidor usando a biblioteca `github.com/gorilla/websocket`.

- **Aprendizado de Go:** Foco no aprimoramento e entendimento dos conceitos fundamentais da linguagem Go através do desenvolvimento prático de um cliente de chat.

## Instalação e Uso
Para iniciar o client:

1. Clone o repositório:

```bash
git clone https://github.com/8january/terminal_chat_client_golang.git
cd terminal_chat_client_golang
```

2. Execute o client:
```bash
go run main.go
```

Para uilizar o servidor, veja [github.com/8january/terminal_chat_server_golang](https://github.com/8january/terminal_chat_server_golang)

### Flags Disponíveis
-addr: Especifica o endereço do servidor de chat. Por padrão, utiliza localhost:8080.

-server: Especifica o servidor a ser conectado. As opções disponíveis são:

default: Conectar ao servidor da flag -addr (localhost:8080 por padrão).

render: Conectar ao servidor hospedado no Render.

### Exemplo de Uso
- Para conectar a um endereço específico:
```bash
go run main.go -addr url_especifica
```

- Para conectar ao servidor local:

```bash
go run main.go -server default
```

- Para conectar ao servidor hospedado no Render:

```bash
go run main.go -server render
```

### Contribuição
Este projeto é experimental e não está planejado para desenvolvimento futuro. Contribuições não são esperadas neste momento.
