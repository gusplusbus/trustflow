defmodule AuthServer.Repo do
  use Ecto.Repo,
    otp_app: :auth_server,
    adapter: Ecto.Adapters.Postgres
end
