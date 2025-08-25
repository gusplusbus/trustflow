import { Paper, Stack, Typography, Button, Chip, Alert } from '@mui/material'
import { useAccount, useChainId, useSwitchChain } from 'wagmi'
import { ConnectButton } from '@rainbow-me/rainbowkit'
import { useProjectWallet } from '../../../hooks/project'

function short(addr?: `0x${string}`) {
  return addr ? `${addr.slice(0, 6)}…${addr.slice(-4)}` : ''
}

export default function AttachCard({ projectId }: { projectId: string }) {
  const { isConnected, address } = useAccount()
  const chainId = useChainId()
  const { switchChainAsync, isPending } = useSwitchChain()
  const { wallet, attach, detach } = useProjectWallet(projectId)

  const expected =
    (import.meta.env.VITE_BLOCKCHAIN_NETWORK as string) === 'sepolia'
      ? 11155111
      : 31337

  const wrongNetwork = isConnected && chainId !== expected

  return (
    <Paper sx={{ p: 2 }}>
      <Stack spacing={1}>
        <Typography variant="h6">Attach wallet to this project</Typography>

        {/* Show current attachment (from localStorage) */}
        {wallet ? (
          <Stack direction="row" spacing={1} alignItems="center">
            <Chip label={`Attached: ${short(wallet.address)}`} color="success" />
            <Chip label={`Chain: ${wallet.chainId}`} size="small" />
            <Button onClick={detach} size="small" variant="outlined">Detach</Button>
          </Stack>
        ) : (
          <Typography color="text.secondary">No wallet attached yet.</Typography>
        )}

        {/* Connect step */}
        {!isConnected && (
          <>
            <Typography>Connect a wallet to attach it.</Typography>
            <ConnectButton showBalance={false} />
          </>
        )}

        {/* Network guard */}
        {wrongNetwork && (
          <Alert
            severity="warning"
            action={
              <Button
                onClick={() => switchChainAsync({ chainId: expected })}
                disabled={isPending}
                size="small"
                variant="contained"
              >
                Switch
              </Button>
            }
          >
            Wrong network. Expected {expected === 31337 ? 'Local (31337)' : 'Sepolia (11155111)'}.
          </Alert>
        )}

        {/* Attach action */}
        {isConnected && !wrongNetwork && (
          <Stack direction="row" spacing={1} alignItems="center">
            <Chip label={`Connected: ${short(address as any)}`} />
            <Chip label={`Chain: ${chainId}`} size="small" />
            <Button
              variant="contained"
              onClick={() => attach(address as `0x${string}`, chainId)}
            >
              Attach this wallet
            </Button>
          </Stack>
        )}
      </Stack>
    </Paper>
  )
}
