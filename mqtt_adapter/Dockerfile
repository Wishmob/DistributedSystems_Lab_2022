# The base go-image
FROM golang:1.18-alpine

# Create a directory for the app
RUN mkdir /app

# Copy all files from the current directory to the app directory
COPY . /app

# Set working directory
WORKDIR /app

RUN mkdir /logs
# Run command as described:
# go build will build an executable file named mqtt_adapter in the current directory
RUN go build -o mqtt_adapter .

# Run the mqtt_adapter executable
CMD [ "/app/mqtt_adapter" ]