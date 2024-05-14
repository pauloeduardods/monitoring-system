# Camera Monitor

Este projeto é um sistema de monitoramento de câmeras em Go que grava vídeos de uma webcam e os salva no Amazon S3. Ele foi projetado para rodar em um Raspberry Pi, mas pode ser executado em qualquer sistema Linux com suporte a Go e OpenCV.

## Pré-requisitos

- Go (Golang) instalado
- OpenCV instalado
- Biblioteca `gocv` instalada
- AWS CLI configurado com credenciais válidas

## Instalação

1. **Clone o repositório:**

   ```sh
   git clone https://github.com/seu-usuario/go-camera.git
   cd go-camera
   ```

2. **Instale as dependências:**

   ```sh
    make deps
   ```

3. **Configure as variáveis de ambiente:**
   Crie um arquivo .env na raiz do projeto com o seguinte conteúdo:

   ```sh
   AWS_REGION=us-east-1
   AWS_BUCKET=nome-do-seu-bucket
   ```

4. **Execute o projeto:**

   ```sh
   make run
   ```

## Contribuição

    1.	Fork este repositório
    2.	Crie uma nova branch (git checkout -b feature/nova-feature)
    3.	Commit suas alterações (git commit -m 'Adiciona nova feature')
    4.	Push para a branch (git push origin feature/nova-feature)
    5.	Crie um novo Pull Request
