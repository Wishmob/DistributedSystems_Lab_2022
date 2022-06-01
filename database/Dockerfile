# The base go-image
FROM golang:1.18-alpine

# Create a directory for the app
RUN mkdir /app

# Copy all files from the current directory to the app directory
COPY . /app

# Set working directory
WORKDIR /app

# Run command as described:
# go build will build an executable file named database_server in the current directory
RUN go build -o database_server .

# Run the database_server executable
CMD [ "/app/database_server" ]