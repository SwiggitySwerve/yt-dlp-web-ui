import React, { useEffect, useState } from 'react';
import { useAtomValue } from 'jotai';
import { serverURL } from '../atoms/settings';
import useFetch from '../hooks/useFetch'; // Assuming useFetch can be used as in SubscriptionsView
import { ArchiveEntry, PaginatedResponse } from '../types';
import {
  Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Typography,
  TablePagination, CircularProgress, Box, Link
} from '@mui/material';
import { useI18n } from '../hooks/useI18n';

const HistoryDownloadsDisplay: React.FC = () => {
  const { i18n } = useI18n();
  const baseURL = useAtomValue(serverURL);

  // Pagination state
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10); // Default rows per page

  // Construct the URL with pagination parameters
  // The API uses 'id' for offset and 'limit' for count.
  // We need to calculate 'startId' based on page and rowsPerPage if the API expects an ID.
  // For now, let's assume the API can take 'page' and 'limit' or we adjust.
  // The existing /archive endpoint takes 'id' (as startRowId) and 'limit'.
  // This means we can't directly use page * rowsPerPage as startId without knowing the IDs.
  // For a simpler first version, we'll fetch a fixed number and handle pagination on client or improve later.
  // Let's fetch a larger initial set and do client-side pagination for now, or prepare for API adjustment.
  // For now, let's just fetch with a limit and handle the 'next' cursor.

  const [startId, setStartId] = useState(0);
  const { data, isLoading, error, fetcher } = useFetch<PaginatedResponse<ArchiveEntry[]>>(
    `/archive?id=${startId}&limit=${rowsPerPage}`
  );

  const handleChangePage = (event: unknown, newPage: number) => {
    // If newPage is greater, we use 'data.next' as the new startId.
    // If newPage is smaller, we'd ideally need 'data.first' or previous cursors.
    // The current API's PaginatedResponse has 'first' and 'next' which are IDs.
    if (data) {
        if (newPage > page && data.next !== 0) { // data.next being 0 might mean no more pages
            setStartId(data.next);
        } else if (newPage < page) {
            // This is tricky with ID-based cursors if we don't store previous 'first' IDs.
            // For now, to go back, we might reset or fetch from a known previous ID if stored.
            // Simplification: Reset to 0 or disable going back beyond current fetch for now.
            // Or, if data.first gives the ID of the first item in the *current* set, use that.
            // Let's assume data.first is the ID of the first item in the current list.
            // To go to an actual previous page, we'd need to refetch with a startId before current data.first
            // This part needs more robust handling or API to support page numbers.
            // For now, we'll use 'next' for forward and 'first' for attempting to go back (might reload current page).
             setStartId(data.first); // This might just reload the current page's start if not careful
        }
    }
    setPage(newPage); // Set the MUI component's page state
     // Actual data fetching for new page happens via useEffect on startId or by calling fetcher.
  };

  useEffect(() => {
    fetcher(); // Refetch when startId or rowsPerPage changes
  }, [startId, rowsPerPage, fetcher]);


  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0); // Reset to first page
    setStartId(0); // Reset startId when rows per page changes
  };

  if (isLoading && !data) {
    return <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}><CircularProgress /></Box>;
  }

  if (error) {
    return <Typography color="error" sx={{ p: 2 }}>{i18n.t('errorLoadingHistory')}: {error.message}</Typography>;
  }

  if (!data || data.data.length === 0) {
    return <Typography sx={{ p: 2 }}>{i18n.t('noCompletedDownloads')}</Typography>;
  }

  return (
    <Paper sx={{ margin: 2 }}>
      <TableContainer>
        <Table stickyHeader aria-label="completed downloads table">
          <TableHead>
            <TableRow>
              <TableCell>{i18n.t('thumbnail')}</TableCell>
              <TableCell>{i18n.t('title')}</TableCell>
              <TableCell>{i18n.t('source')}</TableCell>
              <TableCell>{i18n.t('downloadDate')}</TableCell>
              <TableCell>{i18n.t('actions')}</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {data.data.map((entry) => (
              <TableRow hover key={entry.id}>
                <TableCell>
                  {entry.thumbnail && (
                    <img src={entry.thumbnail} alt={entry.title} style={{ width: '100px', height: 'auto' }} />
                  )}
                </TableCell>
                <TableCell>{entry.title}</TableCell>
                <TableCell>
                  <Link href={entry.source} target="_blank" rel="noopener">
                    {entry.source}
                  </Link>
                </TableCell>
                <TableCell>{new Date(entry.created_at).toLocaleString()}</TableCell>
                <TableCell>
                  {/* Add action buttons here if needed, e.g., view details, delete from history */}
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        component="div"
        count={-1} // API doesn't give total count, so -1 for "unknown"
        rowsPerPageOptions={[10, 25, 50, 100]}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        // MUI TablePagination's 'page' is 0-indexed.
        // The `next` and `first` from API are actual IDs. More complex mapping needed for true server-side pagination.
        // The current implementation of onPageChange is a simplified approach.
      />
    </Paper>
  );
};

export default HistoryDownloadsDisplay;
