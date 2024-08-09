# Imagen base
FROM golang:1.20

# Establecer el directorio de trabajo
WORKDIR /app

# Copiar los archivos del módulo y descargar las dependencias
COPY go.mod ./
COPY go.sum ./
RUN go mod download

# Copiar el resto del código fuente
COPY . .

# Compilar la aplicación
RUN go build -o /mi-servidor

# Comando para ejecutar el servidor
CMD [ "/mi-servidor" ]

