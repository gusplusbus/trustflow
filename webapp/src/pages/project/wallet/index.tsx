import { useParams } from 'react-router-dom'
import AttachCard from './AttachCard'
import ChainGuard from './chainGuard'

export default function Wallet() {
  const { id } = useParams<{ id: string }>()
  if (!id) return <div style={{ padding: 16 }}>No project id.</div>

  return (
    <div style={{ padding: 16 }}>
      <h2>Project Wallet</h2>
      {/* ChainGuard only blocks actions; attach card still shows status */}
      <ChainGuard>
        <AttachCard projectId={id} />
      </ChainGuard>
    </div>
  )
}
