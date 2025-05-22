import React, { useEffect, useState } from 'react'; // Ensure React is imported
import { useAtomValue } from 'jotai';
import { serverURL } from '../atoms/settings'; 
import useFetch from '../hooks/useFetch';
import { ArchiveEntry, PaginatedResponse } from '../types'; 
// Updated imports for MUI controls
import { Container, Grid, Typography, CircularProgress, Box, TablePagination, Select, MenuItem, FormControl, InputLabel, TextField, Button } from '@mui/material';
import { useI18n } from '../hooks/useI18n';
import MediaCard from '../components/media/MediaCard'; 

export default function MediaView() {
  const { i18n } = useI18n();
  // const baseURL = useAtomValue(serverURL); // Not directly used if useFetch handles base URL

  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10); 
  const [startId, setStartId] = useState(0); 

  // State for Sort/Filter values
  const [sortOption, setSortOption] = useState('');
  const [filterUploader, setFilterUploader] = useState('');

  // 1. Update useFetch URL construction
  const queryParams = new URLSearchParams();
  queryParams.append('id', String(startId));
  queryParams.append('limit', String(rowsPerPage));
  if (sortOption) {
    queryParams.append('sort_by', sortOption);
  }
  if (filterUploader) {
    queryParams.append('filter_uploader', filterUploader);
  }

  const { data: archiveData, isLoading, error, fetcher } = useFetch<PaginatedResponse<ArchiveEntry[]>>(
    `/archive?${queryParams.toString()}`
  );

  // 3. Update useEffect for fetcher
  useEffect(() => {
    console.log("MediaView: Fetching archive data due to dependency change.", { startId, rowsPerPage, sortOption, filterUploader });
    fetcher();
  }, [startId, rowsPerPage, sortOption, filterUploader, fetcher]); // Added sortOption, filterUploader

  const handleChangePage = (event: unknown, newPage: number) => {
    console.log("MediaView: handleChangePage", { newPage, currentPage: page });
    if (archiveData) {
      if (newPage > page && archiveData.next !== 0) {
        setStartId(archiveData.next);
      } else if (newPage < page && archiveData.first !== startId) { // Avoid re-fetching same page if 'first' is current startId
        // This logic for 'previous' needs to be robust if API provides 'prev_cursor' or similar.
        // For now, using 'first' might take you to the beginning of the current set or the absolute first page.
        // If your API's 'first' cursor means "ID of the first item in the current list", 
        // then to go "back" you'd need a different mechanism or for startId to be 0.
        // Setting to 0 often means "fetch the very first page".
        // For now, let's assume going back means resetting to the first page of the current filter/sort.
        // This part may need server-side support for true previous page cursors.
        // A simple approach for "previous" might just be to set startId to 0 if newPage is 0.
        if (newPage === 0) {
            setStartId(0);
        } else {
            // Complex "previous" logic with cursors is hard without API support.
            // For now, this might re-fetch from 'archiveData.first' or do nothing if it's the same.
            setStartId(archiveData.first); 
        }
      }
    }
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    console.log("MediaView: handleChangeRowsPerPage", { newRowsPerPage: event.target.value });
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
      
      <Box mb={3} display="flex" gap={2} alignItems="center" flexWrap="wrap">
        <FormControl sx={{ minWidth: 180 }} size="small">
          <InputLabel id="sort-by-label">{i18n.t('sortByLabel')}</InputLabel>
          <Select
            labelId="sort-by-label"
            label={i18n.t('sortByLabel')}
            value={sortOption}
            onChange={(e) => {
              console.log("MediaView: Sort option changed", { newSortOption: e.target.value });
              setSortOption(e.target.value);
              // Optional: Apply immediately or wait for "Apply" button
              // If applying immediately:
              // setPage(0);
              // setStartId(0);
            }}
          >
            <MenuItem value="">{i18n.t('noneSortOption')}</MenuItem>
            <MenuItem value="title_asc">{i18n.t('titleAscSortOption')}</MenuItem>
            <MenuItem value="title_desc">{i18n.t('titleDescSortOption')}</MenuItem>
            <MenuItem value="date_asc">{i18n.t('dateAscSortOption')}</MenuItem>
            <MenuItem value="date_desc">{i18n.t('dateDescSortOption')}</MenuItem>
            <MenuItem value="duration_asc">{i18n.t('durationAscSortOption')}</MenuItem>
            <MenuItem value="duration_desc">{i18n.t('durationDescSortOption')}</MenuItem>
          </Select>
        </FormControl>

        <TextField
          label={i18n.t('filterByUploaderLabel')}
          variant="outlined"
          size="small"
          value={filterUploader}
          onChange={(e) => {
            console.log("MediaView: Filter uploader changed", { newFilterUploader: e.target.value });
            setFilterUploader(e.target.value);
            // Optional: Apply immediately or wait for "Apply" button
          }}
          sx={{ minWidth: 200 }}
        />
        
        {/* 2. Update "Apply" Button's onClick handler */}
        <Button variant="contained" onClick={() => {
            console.log("MediaView: Apply Filters/Sort Clicked", { sortOption, filterUploader, currentPage: page, currentStartId: startId });
            setPage(0); 
            setStartId(0); 
            // The useEffect will now trigger the fetch due to sortOption/filterUploader being in its deps,
            // and startId/page being reset.
          }}>
          {i18n.t('applyFiltersButtonLabel')}
        </Button>
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
        rowsPerPageOptions={[10, 25, 50, 100]} 
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
        sx={{ mt: 2 }}
      />
    </Container>
  );
}
