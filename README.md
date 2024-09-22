# Camera Monitor

Este projeto é um sistema de monitoramento de câmeras em Go que grava vídeos de uma webcam utilizando OpenCV. Ele foi projetado para rodar em um Raspberry Pi, mas pode ser executado em qualquer sistema Linux ou Mac com suporte a Go e OpenCV.

## Pré-requisitos

- Go (Golang) instalado
- OpenCV instalado
- Biblioteca `gocv` instalada

## Instalação

1. **Clone o repositório:**

   ```sh
   git clone https://github.com/pauloeduardods/monitoring-system
   cd monitoring-system
   ```

2. **Instale as dependências:**

   ```sh
    make deps
   ```

3.**Execute o projeto (sem Docker):**

   ```sh
   make run
   ```

## Executando com Docker

1. Build da imagem Docker

Caso prefira rodar o projeto dentro de um container Docker, você pode buildar a imagem da seguinte forma:

   ```sh
   docker build -t monitoring-system .
   ```

2. Executando o container Docker com acesso à câmera

Para rodar o container com acesso aos dispositivos de câmera, use a flag --device para mapear os dispositivos de vídeo para dentro do contêiner. Geralmente, as câmeras no Linux estão mapeadas em /dev/video\*.
Exemplo para rodar o container com uma câmera conectada:

   ```sh
   docker run --rm -it \
    --device=/dev/video0:/dev/video0 \
    -p 4000:4000 \
    monitoring-system
   ```

Se houver mais câmeras conectadas, você pode mapear mais dispositivos conforme necessário, por exemplo, adicionando --device=/dev/video1:/dev/video1. 3. Redirecionar todas as câmeras (opcional)

Caso você queira dar ao contêiner acesso a todos os dispositivos do sistema, você pode utilizar o modo --privileged:

   ```sh
   docker run --rm -it --privileged -p 4000:4000 monitoring-system
   ```

Isso garante que todos os dispositivos sejam acessíveis dentro do contêiner.
