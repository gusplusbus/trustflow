FROM elixir:1.16

# Install Hex + Rebar
RUN mix local.hex --force && \
    mix local.rebar --force

# Install system deps: nodejs, npm, postgresql-client, inotify-tools
RUN apt-get update && apt-get install -y \
    nodejs \
    npm \
    postgresql-client \
    inotify-tools \
 && apt-get clean

# Set working directory
WORKDIR /app

# Copy project files
COPY . .

# Install Elixir dependencies
RUN mix deps.get
RUN mix deps.compile

# Compile the Phoenix app
RUN mix compile

EXPOSE 4000
CMD ["mix", "phx.server"]
