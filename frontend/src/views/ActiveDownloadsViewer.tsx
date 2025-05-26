import React, { useEffect, useState } from 'react';
import {
  Box,
  Typography,
  Button,
  CircularProgress,
  Paper,
  Grid,
  Card,
  CardContent,
  CardActions,
  LinearProgress,
  IconButton,
  Tooltip,
} from '@mui/material';
import CancelIcon from '@mui/icons-material/Cancel';
import ClearIcon from '@mui/icons-material/Clear';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import { useAtomValue } from 'jotai';
import { activeDownloadsState } from 'frontend/src/atoms/downloads';
import { useRPC } from 'frontend/src/hooks/useRPC';
import { useToast } from 'frontend/src/hooks/toast';
import { useI18n } from 'frontend/src/hooks/useI18n';
import { RPCResult } from 'frontend/src/types';

const ActiveDownloadsViewer: React.FC = () => {
  const { i18n } = useI18n();
  const rpcClient = useRPC();
  const { pushMessage } = useToast();
  const downloads = useAtomValue(activeDownloadsState);

  // Categorize downloads
  const pendingDownloads = downloads.filter(d => d.progress?.status === 0 || d.progress?.status === 'pending');
  const activeDownloading = downloads.filter(d => d.progress?.status === 1 || d.progress?.status === 'downloading');
  const erroredDownloads = downloads.filter(d => d.progress?.status === 3 || d.progress?.status === 'errored');

  const handleCancel = async (id: string) => {
    try {
      await rpcClient.kill(id);
      pushMessage(i18n.t('downloadCancelledMessage', { id }), 'success');
    } catch (error: any) {
      pushMessage(i18n.t('errorCancellingDownload', { id, message: error.message }), 'error');
    }
  };

  const handleClearErrored = async (id: string) => {
    try {
      await rpcClient.clear(id);
      pushMessage(i18n.t('downloadClearedMessage', { id }), 'success');
    } catch (error: any) {
      pushMessage(i18n.t('errorClearingDownload', { id, message: error.message }), 'error');
    }
  };

  return (
    <Paper sx={{ p: 2 }}>
      <Typography variant="h5" gutterBottom>
        {i18n.t('ongoingDownloadsTitle')}
      </Typography>

      {/* Pending Downloads Section */}
      <Box mb={3}>
        <Typography variant="h6" gutterBottom>
          {i18n.t('pendingDownloadsTitle')}
        </Typography>
        {pendingDownloads.length === 0 ? (
          <Typography>{i18n.t('noPendingDownloads')}</Typography>
        ) : (
          <Grid container spacing={2}>
            {pendingDownloads.map((download) => (
              <Grid item xs={12} sm={6} md={4} key={download.id}>
                <Card>
                  <CardContent>
                    <Typography variant="subtitle1">{download.name || download.title || download.id}</Typography>
                    <Typography variant="body2" color="textSecondary">
                      {i18n.t('downloadStatusPending')}
                    </Typography>
                  </CardContent>
                  <CardActions>
                    <Tooltip title={i18n.t('cancelDownloadButtonTooltip')}>
                      <IconButton onClick={() => handleCancel(download.id)} size="small">
                        <CancelIcon />
                      </IconButton>
                    </Tooltip>
                  </CardActions>
                </Card>
              </Grid>
            ))}
          </Grid>
        )}
      </Box>

      {/* Active Downloads Section */}
      <Box mb={3}>
        <Typography variant="h6" gutterBottom>
          {i18n.t('activeDownloadsTitle')}
        </Typography>
        {activeDownloading.length === 0 ? (
          <Typography>{i18n.t('noActiveDownloads')}</Typography>
        ) : (
          <Grid container spacing={2}>
            {activeDownloading.map((download) => (
              <Grid item xs={12} sm={6} md={4} key={download.id}>
                <Card>
                  <CardContent>
                    <Typography variant="subtitle1">{download.name || download.title || download.id}</Typography>
                    <Typography variant="body2" color="textSecondary">
                      {i18n.t('downloadStatusDownloading')}
                    </Typography>
                    {download.progress?.percentage && (
                      <Box sx={{ display: 'flex', alignItems: 'center', mt: 1 }}>
                        <Box sx={{ width: '100%', mr: 1 }}>
                          <LinearProgress variant="determinate" value={parseFloat(download.progress.percentage)} />
                        </Box>
                        <Box sx={{ minWidth: 35 }}>
                          <Typography variant="body2" color="textSecondary">{`${download.progress.percentage}%`}</Typography>
                        </Box>
                      </Box>
                    )}
                    <Typography variant="caption" display="block" sx={{ mt: 0.5 }}>
                      {i18n.t('downloadProgressLabel', { 
                        percentage: download.progress?.percentage || 'N/A', 
                        speed: download.progress?.speed || 'N/A', 
                        eta: download.progress?.eta || 'N/A' 
                      })}
                    </Typography>
                  </CardContent>
                  <CardActions>
                    <Tooltip title={i18n.t('cancelDownloadButtonTooltip')}>
                      <IconButton onClick={() => handleCancel(download.id)} size="small">
                        <CancelIcon />
                      </IconButton>
                    </Tooltip>
                  </CardActions>
                </Card>
              </Grid>
            ))}
          </Grid>
        )}
      </Box>

      {/* Errored Downloads Section */}
      <Box>
        <Typography variant="h6" gutterBottom>
          {i18n.t('erroredDownloadsTitle')}
        </Typography>
        {erroredDownloads.length === 0 ? (
          <Typography>{i18n.t('noErroredDownloads')}</Typography>
        ) : (
          <Grid container spacing={2}>
            {erroredDownloads.map((download) => (
              <Grid item xs={12} sm={6} md={4} key={download.id}>
                <Card>
                  <CardContent>
                    <Box sx={{ display: 'flex', alignItems: 'center' }}>
                      <ErrorOutlineIcon color="error" sx={{ mr: 1 }} />
                      <Typography variant="subtitle1">{download.name || download.title || download.id}</Typography>
                    </Box>
                    <Typography variant="body2" color="error" sx={{ mt: 1 }}>
                      {i18n.t('downloadStatusErrored')}
                    </Typography>
                    {download.progress?.error && (
                      <Typography variant="caption" color="error" display="block">
                        {download.progress.error}
                      </Typography>
                    )}
                  </CardContent>
                  <CardActions>
                    <Tooltip title={i18n.t('clearErroredButtonTooltip')}>
                      <IconButton onClick={() => handleClearErrored(download.id)} size="small">
                        <ClearIcon />
                      </IconButton>
                    </Tooltip>
                  </CardActions>
                </Card>
              </Grid>
            ))}
          </Grid>
        )}
      </Box>
    </Paper>
  );
};

export default ActiveDownloadsViewer;
