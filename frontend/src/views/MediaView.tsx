import React, { useEffect, useState } from 'react'; 
import { useAtomValue } from 'jotai';
import { serverURL } from '../atoms/settings'; 
import useFetch from '../hooks/useFetch';
import { ArchiveEntry, PaginatedResponse } from '../types'; 
import { Container, Grid, Typography, CircularProgress, Box, TablePagination, Select, MenuItem, FormControl, InputLabel, TextField, Button } from '@mui/material';
import { useI18n } from '../hooks/useI18n';
import MediaCard from '../components/media/MediaCard'; 

export default function MediaView() {
  const { i18n } = useI18n();

  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10); 
  const [startId, setStartId] = useState(0); 

  // State for Sort/Filter values
  const [sortOption, setSortOption] = useState('');
  const [filterUploader, setFilterUploader] = useState('');
  const [filterFormat, setFilterFormat] = useState(''); 
  const [filterMinDuration, setFilterMinDuration] = useState(''); 
  const [filterMaxDuration, setFilterMaxDuration] = useState(''); 
  const [searchQuery, setSearchQuery] = useState(''); // Step 1: Add State Variable

  // Update useFetch URL Construction
  const queryParams = new URLSearchParams();
  queryParams.append('id', String(startId));
  queryParams.append('limit', String(rowsPerPage));
  if (sortOption) {
    queryParams.append('sort_by', sortOption);
  }
  if (filterUploader) {
    queryParams.append('filter_uploader', filterUploader);
  }
  if (filterFormat) {
    queryParams.append('filter_format', filterFormat);
  }
  if (filterMinDuration) {
    queryParams.append('filter_min_duration', filterMinDuration);
  }
  if (filterMaxDuration) {
    queryParams.append('filter_max_duration', filterMaxDuration);
  }
  if (searchQuery) { // Step 3: Add search_query to URL
    queryParams.append('search_query', searchQuery);
  }

  const { data: archiveData, isLoading, error, fetcher } = useFetch<PaginatedResponse<ArchiveEntry[]>>(
    `/archive?${queryParams.toString()}`
  );

  // Update useEffect for fetcher
  useEffect(() => {
    console.log("MediaView: Fetching archive data due to dependency change.", { 
        startId, rowsPerPage, sortOption, filterUploader, 
        filterFormat, filterMinDuration, filterMaxDuration, searchQuery // Added searchQuery
    });
    fetcher();
  }, [startId, rowsPerPage, sortOption, filterUploader, filterFormat, filterMinDuration, filterMaxDuration, searchQuery, fetcher]); // Step 4: Add searchQuery to deps

  const handleChangePage = (event: unknown, newPage: number) => {
    console.log("MediaView: handleChangePage", { newPage, currentPage: page });
    if (archiveData) {
      if (newPage > page && archiveData.next !== 0) {
        setStartId(archiveData.next);
      } else if (newPage < page && archiveData.first !== startId) {
        if (newPage === 0) {
            setStartId(0);
        } else {
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
      
      {/* Filter and Sort Controls Box */}
      <Box mb={3} display="flex" gap={2} alignItems="center" flexWrap="wrap">
        {/* Step 2: Add Search Bar UI Element */}
        <TextField
          label={i18n.t('searchMediaLabel')}
          variant="outlined"
          size="small"
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          onKeyDown={(e) => {
              if (e.key === 'Enter') {
                  console.log("MediaView: Search on Enter:", searchQuery);
                  setPage(0); 
                  setStartId(0);
              }
          }}
          sx={{ width: { xs: '100%', sm: 250, md: 300 }, mr: {sm: 'auto'} }} // Takes more space on xs, specific width on sm+
        />

        <FormControl sx={{ minWidth: 180 }} size="small">
          <InputLabel id="sort-by-label">{i18n.t('sortByLabel')}</InputLabel>
          <Select
            labelId="sort-by-label"
            label={i18n.t('sortByLabel')}
            value={sortOption}
            onChange={(e) => {
              setSortOption(e.target.value);
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
            setFilterUploader(e.target.value);
          }}
          sx={{ minWidth: 200 }}
        />

        <FormControl sx={{ minWidth: 150 }} size="small">
          <InputLabel id="filter-format-label">{i18n.t('filterByFormatLabel')}</InputLabel>
          <Select
            labelId="filter-format-label"
            label={i18n.t('filterByFormatLabel')}
            value={filterFormat}
            onChange={(e) => setFilterFormat(e.target.value)}
          >
            <MenuItem value="">{i18n.t('allFormatsOption')}</MenuItem>
            <MenuItem value="mp4">MP4</MenuItem>
            <MenuItem value="webm">WEBM</MenuItem>
            <MenuItem value="mkv">MKV</MenuItem>
          </Select>
        </FormControl>

        <TextField
          label={i18n.t('filterMinDurationLabel')}
          variant="outlined"
          size="small"
          type="number"
          value={filterMinDuration}
          onChange={(e) => setFilterMinDuration(e.target.value)}
          sx={{ width: 180 }}
          InputProps={{ inputProps: { min: 0 } }}
        />

        <TextField
          label={i18n.t('filterMaxDurationLabel')}
          variant="outlined"
          size="small"
          type="number"
          value={filterMaxDuration}
          onChange={(e) => setFilterMaxDuration(e.target.value)}
          sx={{ width: 180 }}
          InputProps={{ inputProps: { min: 0 } }}
        />
        
        <Button variant="contained" onClick={() => {
            console.log("MediaView: Apply Filters/Sort Clicked", { 
                sortOption, filterUploader, filterFormat, 
                filterMinDuration, filterMaxDuration, searchQuery, // Step 5: Update console.log
                currentPage: page, currentStartId: startId 
            });
            setPage(0); 
            setStartId(0); 
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
        count={-1} 
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
