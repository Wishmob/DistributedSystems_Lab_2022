# The base go-image
FROM golang:1.18-alpine

# Create a directory for the app
RUN mkdir /app

# Copy all files from the current directory to the app directory
COPY . /app

# Set working directory
WORKDIR /app

# Create folder for logs
RUN mkdir /logs

# Run command as described:
# go build will build an executable file named iot_gateway in the current directory
RUN go build -o iot_gateway .

# Run the iot_gateway executable
CMD [ "/app/iot_gateway" ]