import { useAccount, useChainId, useSwitchChain } from 'wagmi'

export default function ChainGuard({ children }: { children: React.ReactNode }) {
  const { isConnected } = useAccount()
  const chainId = useChainId()
  const { switchChainAsync, isPending } = useSwitchChain()

  const expected = import.meta.env.VITE_BLOCKCHAIN_NETWORK === 'local' ? 31337 : 11155111

  if (!isConnected) return <>{children}</>
  if (chainId !== expected) {
    return (
      <div style={{ padding: 16 }}>
        <p>You’re on the wrong network.</p>
        <button disabled={isPending} onClick={() => switchChainAsync({ chainId: expected })}>
          Switch network
        </button>
      </div>
    )
  }
  return <>{children}</>
}
