import Config

# Enable PHX server if set
if System.get_env("PHX_SERVER") do
  config :auth_server, AuthServerWeb.Endpoint, server: true
end

# Configure the database URL and Repo for ALL environments
database_url =
  System.get_env("DATABASE_URL") ||
    raise """
    environment variable DATABASE_URL is missing.
    For example: ecto://USER:PASS@HOST/DATABASE
    """

maybe_ipv6 = if System.get_env("ECTO_IPV6") in ~w(true 1), do: [:inet6], else: []

config :auth_server, AuthServer.Repo,
  url: database_url,
  pool_size: String.to_integer(System.get_env("POOL_SIZE") || "10"),
  socket_options: maybe_ipv6

if config_env() != :prod do
  config :auth_server, AuthServerWeb.Endpoint,
    http: [ip: {0, 0, 0, 0}, port: 4000]
end

# Only prod-specific config
if config_env() == :prod do
  secret_key_base =
    System.get_env("SECRET_KEY_BASE") ||
      raise """
      environment variable SECRET_KEY_BASE is missing.
      You can generate one by calling: mix phx.gen.secret
      """

  host = System.get_env("PHX_HOST") || "example.com"
  port = String.to_integer(System.get_env("PORT") || "4000")

  config :auth_server, :dns_cluster_query, System.get_env("DNS_CLUSTER_QUERY")

  config :auth_server, AuthServerWeb.Endpoint,
    url: [host: host, port: 443, scheme: "https"],
    http: [
      ip: {0, 0, 0, 0, 0, 0, 0, 0},
      port: port
    ],
    secret_key_base: secret_key_base
end
