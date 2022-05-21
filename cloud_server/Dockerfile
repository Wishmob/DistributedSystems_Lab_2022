# The base go-image
FROM golang:1.18-alpine

# Create a directory for the app
RUN mkdir /app

# Copy all files from the current directory to the app directory
COPY . /app

# Set working directory
WORKDIR /app

# Run command as described:
# go build will build an executable file named cloud_server in the current directory
RUN go build -o cloud_server .

# Run the cloud_server executable
CMD [ "/app/cloud_server" ]