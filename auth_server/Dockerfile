# Dockerfile
FROM elixir:1.16

# Install Hex + Rebar
RUN mix local.hex --force && \
    mix local.rebar --force

# Install Node.js for assets (optional for LiveView)
RUN apt-get update && apt-get install -y nodejs npm postgresql-client

WORKDIR /app

COPY . .

# Install deps
RUN mix deps.get
RUN mix deps.compile

# Compile the app
RUN mix compile

EXPOSE 4000
CMD ["mix", "phx.server"]
