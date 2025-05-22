import { useAtom, useAtomValue } from 'jotai'
import { useEffect, useState } from 'react'
import { loadingDownloadsState } from '../atoms/downloads'
import { listViewState } from '../atoms/settings'
import { loadingAtom } from '../atoms/ui'
// import DownloadsGridView from './DownloadsGridView' // No longer directly used here
// import DownloadsTableView from './DownloadsTableView' // No longer directly used here
import { Tabs, Tab, Box, Typography } from '@mui/material'
import ActiveDownloadsDisplay from './ActiveDownloadsDisplay'
import PendingDownloadsDisplay from './PendingDownloadsDisplay'
import HistoryDownloadsDisplay from './HistoryDownloadsDisplay'

const Downloads: React.FC = () => {
  // const tableView = useAtomValue(listViewState) // tableView is now used within the display components
  const loadingDownloads = useAtomValue(loadingDownloadsState)
  const [currentTab, setCurrentTab] = useState(0)

  const [isLoading, setIsLoading] = useAtom(loadingAtom)

  useEffect(() => {
    if (loadingDownloads) {
      return setIsLoading(true)
    }
    setIsLoading(false)
  }, [loadingDownloads, isLoading])

  // if (tableView) return <DownloadsTableView />

  // return <DownloadsGridView />

  return (
    <>
      <Box sx={{ borderBottom: 1, borderColor: 'divider', mb: 2 }}>
        <Tabs value={currentTab} onChange={(event, newValue) => setCurrentTab(newValue)} aria-label="download sections">
          <Tab label="Active" />
          <Tab label="Pending" />
          <Tab label="History" />
        </Tabs>
      </Box>
      {currentTab === 0 && <ActiveDownloadsDisplay />}
      {currentTab === 1 && <PendingDownloadsDisplay />}
      {currentTab === 2 && <HistoryDownloadsDisplay />}
    </>
  )
}

export default Downloads