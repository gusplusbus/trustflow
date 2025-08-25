import '@rainbow-me/rainbowkit/styles.css'

import { WagmiProvider } from 'wagmi'
import { sepolia } from 'wagmi/chains'
import { foundry } from 'viem/chains'
import { http, type Transport } from 'viem'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { RainbowKitProvider, getDefaultConfig, ConnectButton } from '@rainbow-me/rainbowkit'

const FOUND_RCP =
  (import.meta.env.VITE_BLOCKCHAIN_RPC as string) ||
  foundry.rpcUrls.default.http[0] ||
  'http://localhost:8545'

const SEP_RCP =
  (import.meta.env.VITE_SEPOLIA_RPC as string) ||
  sepolia.rpcUrls.default.http[0]

const chains = [foundry, sepolia]
const transports: Record<number, Transport> = {
  [foundry.id]: http(FOUND_RCP),
  [sepolia.id]: http(SEP_RCP),
}

const config = getDefaultConfig({
  appName: 'Trustflow',
  projectId: (import.meta.env.VITE_BLOCKCHAIN_WC_PROJECT_ID as string) || 'dev',
  chains,
  transports,
})

const queryClient = new QueryClient()

export function WalletShell({ children }: { children: React.ReactNode }) {
  return (
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <RainbowKitProvider>
          <div style={{ display: 'flex', justifyContent: 'flex-end', padding: 12 }}>
            <ConnectButton showBalance={false} />
          </div>
          {children}
        </RainbowKitProvider>
      </QueryClientProvider>
    </WagmiProvider>
  )
}
