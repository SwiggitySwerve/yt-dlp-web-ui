import DeleteIcon from '@mui/icons-material/Delete'
import EditIcon from '@mui/icons-material/Edit'
import {
  Box,
  Button,
  Container,
  Paper,
  Table, TableBody, TableCell, TableContainer,
  TableHead, TablePagination, TableRow,
  Typography, // Added
  Paper // Added
} from '@mui/material'
import { matchW } from 'fp-ts/lib/Either'
import { pipe } from 'fp-ts/lib/function'
import { useAtomValue } from 'jotai'
import { useState, useTransition } from 'react'
import { serverURL } from '../atoms/settings'
import LoadingBackdrop from '../components/LoadingBackdrop'
import NoSubscriptions from '../components/subscriptions/NoSubscriptions'
import SubscriptionsDialog from '../components/subscriptions/SubscriptionsDialog'
import SubscriptionsEditDialog from '../components/subscriptions/SubscriptionsEditDialog'
import SubscriptionsSpeedDial from '../components/subscriptions/SubscriptionsSpeedDial'
import ChannelVideosView from '../components/subscriptions/ChannelVideosView'; // Added import
import { useToast } from '../hooks/toast'
import useFetch from '../hooks/useFetch'
import { useI18n } from '../hooks/useI18n'
import { ffetch } from '../lib/httpClient'
import { Subscription } from '../services/subscriptions'
import { PaginatedResponse, YtdlpChannelDump } from '../types' // Added YtdlpChannelDump

const SubscriptionsView: React.FC = () => {
  const { i18n } = useI18n()
  const { pushMessage } = useToast()

  const baseURL = useAtomValue(serverURL)

  const [selectedSubscription, setSelectedSubscription] = useState<Subscription>()
  const [openDialog, setOpenDialog] = useState(false)

  const [startId, setStartId] = useState(0)
  const [limit, setLimit] = useState(9)
  const [page, setPage] = useState(0)

  // State for viewing channel videos
  const [viewingSubscriptionId, setViewingSubscriptionId] = useState<string | null>(null);
  const [channelVideosData, setChannelVideosData] = useState<YtdlpChannelDump | null>(null);
  const [isLoadingChannelVideos, setIsLoadingChannelVideos] = useState<boolean>(false);

  const { data: subs, fetcher: refecth } = useFetch<PaginatedResponse<Subscription[]>>(
    `/subscriptions?id=${startId}&limit=${limit}`
  )

  const [isPending, startTransition] = useTransition()

  const handleViewVideos = async (id: string) => {
    setViewingSubscriptionId(id);
    setIsLoadingChannelVideos(true);
    setChannelVideosData(null); // Clear previous data

    const task = ffetch<YtdlpChannelDump>(`${baseURL}/subscriptions/${id}/videos`);
    const either = await task();

    pipe(
      either,
      matchW(
        (error) => {
          pushMessage(`Error fetching channel videos: ${error.message || error}`, 'error');
          setIsLoadingChannelVideos(false);
          setViewingSubscriptionId(null); // Clear on error
        },
        (data) => {
          setChannelVideosData(data);
          setIsLoadingChannelVideos(false);
        }
      )
    );
  };

  const deleteSubscription = async (id: string) => {
    const task = ffetch<void>(`${baseURL}/subscriptions/${id}`, {
      method: 'DELETE',
    })
    const either = await task()

    pipe(
      either,
      matchW(
        (l) => pushMessage(l, 'error'),
        () => refecth()
      )
    )
  }

  return (
    <>
      <LoadingBackdrop isLoading={!subs || isPending} />

      <SubscriptionsSpeedDial onOpen={() => setOpenDialog(s => !s)} />

      <SubscriptionsEditDialog
        subscription={selectedSubscription}
        onClose={() => {
          setSelectedSubscription(undefined)
          refecth()
        }}
      />
      <SubscriptionsDialog open={openDialog} onClose={() => {
        setOpenDialog(s => !s)
        refecth()
      }} />

      {!subs || subs.data.length === 0 ?
        <NoSubscriptions /> :
        <Container maxWidth="xl" sx={{ mt: 4, mb: 8 }}>
          <Paper sx={{
            p: 2.5,
            display: 'flex',
            flexDirection: 'column',
            minHeight: '80vh',
          }}>
            <TableContainer component={Box}>
              <Table sx={{ minWidth: '100%' }}>
                <TableHead>
                  <TableRow>
                    <TableCell align="left">URL</TableCell>
                    <TableCell align="right">Params</TableCell>
                    <TableCell align="right">{i18n.t('cronExpressionLabel')}</TableCell>
                    <TableCell align="center">Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody sx={{ mb: 'auto' }}>
                  {subs.data.map(x => (
                    <TableRow
                      key={x.id}
                      sx={{ '&:last-child td, &:last-child th': { border: 0 } }}
                    >
                      <TableCell>{x.url}</TableCell>
                      <TableCell align='right'>
                        {x.params}
                      </TableCell>
                      <TableCell align='right'>
                        {x.cron_expression}
                      </TableCell>
                      <TableCell align='center'>
                        <Button
                          variant='contained'
                          size='small'
                          sx={{ mr: 0.5 }}
                          onClick={() => setSelectedSubscription(x)}
                        >
                          <EditIcon />
                        </Button>
                        <Button
                          variant='contained'
                          size='small'
                          onClick={() => startTransition(async () => await deleteSubscription(x.id))}
                        >
                          <DeleteIcon />
                        </Button>
                        <Button
                          variant='outlined'
                          size='small'
                          sx={{ ml: 0.5 }} // Use ml for margin-left if it's to the right of delete
                          onClick={() => handleViewVideos(x.id)}
                        >
                          View Videos
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>

            {/* Display area for channel videos - Replaced with ChannelVideosView */}
            {viewingSubscriptionId && (
              <ChannelVideosView
                isLoading={isLoadingChannelVideos}
                channelData={channelVideosData}
                onClose={() => {
                  setViewingSubscriptionId(null);
                  setChannelVideosData(null);
                }}
              />
            )}
            {/* End display area */}

            <TablePagination
              component="div"
              count={-1}
              page={page}
              onPageChange={(_, p) => {
                if (p < page) {
                  setPage(s => (s - 1 <= 0 ? 0 : s - 1))
                  setStartId(subs.first)
                  return
                }
                setPage(s => s + 1)
                setStartId(subs.next)
              }}
              rowsPerPage={limit}
              rowsPerPageOptions={[9, 10, 25, 50, 100]}
              onRowsPerPageChange={(e) => { setLimit(parseInt(e.target.value)) }}
            />
          </Paper>
        </Container>
      }
    </>
  )
}

export default SubscriptionsView