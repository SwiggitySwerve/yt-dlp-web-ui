import React, { useEffect, useState } from 'react'; // Ensure React is imported
import { useAtomValue } from 'jotai';
import { serverURL } from '../atoms/settings'; // Keep if needed for baseURL with useFetch
import useFetch from '../hooks/useFetch';
import { ArchiveEntry, PaginatedResponse } from '../types'; // ArchiveEntry will be displayed
import { Container, Grid, Typography, CircularProgress, Box, /* Paper, */ TablePagination } from '@mui/material'; // Paper removed as MediaCard handles its own structure
import { useI18n } from '../hooks/useI18n';
import MediaCard from '../components/media/MediaCard'; // Added import

export default function MediaView() {
  const { i18n } = useI18n();
  const baseURL = useAtomValue(serverURL); // Potentially needed if useFetch doesn't auto-prepend it

  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10); // Or a more grid-friendly number like 12 or 24
  const [startId, setStartId] = useState(0); // For cursor pagination if API uses it

  const { data: archiveData, isLoading, error, fetcher } = useFetch<PaginatedResponse<ArchiveEntry[]>>(
    // The /archive endpoint uses 'id' for startRowId and 'limit'
    `/archive?id=${startId}&limit=${rowsPerPage}` 
  );

  useEffect(() => {
    fetcher(); // Refetch when startId or rowsPerPage changes
  }, [startId, rowsPerPage, fetcher]);

  const handleChangePage = (event: unknown, newPage: number) => {
    if (archiveData) {
      if (newPage > page && archiveData.next !== 0) {
        setStartId(archiveData.next);
      } else if (newPage < page) {
        // This needs robust handling for cursor-based pagination's "previous" page.
        // For now, if API provides 'first' as the ID of the first item in current set:
        setStartId(archiveData.first); // This might just reload current page or first page.
        // A true previous page with cursors often requires server-side changes or storing prev cursors.
      }
    }
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(parseInt(event.target.value, 10));
    setPage(0);
    setStartId(0); 
  };

  if (isLoading && !archiveData) {
    return (
      <Container maxWidth="xl" sx={{ mt: 2, mb: 8, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress />
      </Container>
    );
  }

  if (error) {
    return (
      <Container maxWidth="xl" sx={{ mt: 2, mb: 8 }}>
        <Typography color="error">{i18n.t('errorLoadingMedia')}: {error.message}</Typography>
      </Container>
    );
  }

  if (!archiveData || archiveData.data.length === 0) {
    return (
      <Container maxWidth="xl" sx={{ mt: 2, mb: 8 }}>
        <Typography>{i18n.t('noMediaFound')}</Typography>
      </Container>
    );
  }

  return (
    <Container maxWidth="xl" sx={{ mt: 2, mb: 8 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        {i18n.t('mediaPageTitle')}
      </Typography>
      
      {/* Placeholder for Search/Filter/Sort controls - to be added in Step 3 */}
      <Box mb={2}>
        <Typography variant="subtitle1">Search/Filter/Sort Placeholder</Typography>
      </Box>

      <Grid container spacing={2}>
        {archiveData.data.map((entry) => (
          <Grid item xs={12} sm={6} md={4} lg={3} xl={2} key={entry.id}>
            <MediaCard entry={entry} />
          </Grid>
        ))}
      </Grid>
      <TablePagination
        component="div"
        count={-1} // API doesn't give total count
        rowsPerPageOptions={[10, 25, 50, 100]} // Or more grid-friendly numbers
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        sx={{ mt: 2 }}
      />
    </Container>
  );
}
