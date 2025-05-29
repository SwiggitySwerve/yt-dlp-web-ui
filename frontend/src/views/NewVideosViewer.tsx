import React, { useState, useEffect, useCallback } from 'react';
import {
  Container,
  Typography,
  CircularProgress,
  Box,
  Card,
  CardContent,
  CardMedia,
  CardActions,
  Button,
  Grid,
  Tooltip,
  IconButton,
} from '@mui/material';
import CheckIcon from '@mui/icons-material/Check';
import DownloadIcon from '@mui/icons-material/Download';
import PlaylistPlayIcon from '@mui/icons-material/PlaylistPlay'; // For View Channel
import MovieIcon from '@mui/icons-material/Movie'; // Placeholder for missing thumbnail

import { useI18n } from '../hooks/useI18n';
import { useToast } from '../hooks/toast';
import { SubscriptionVideoUpdate } from '../types'; // Assuming type is in 'types/index.ts'
import { ffetch } from '../lib/httpClient';
import { useNavigate } from 'react-router-dom';
import { useAtomValue } from 'jotai';
import { serverURL } from '../atoms/settings';
import { formatDate } from '../utils'; // Assuming formatDate exists

const NewVideosViewer: React.FC = () => {
  const { i18n } = useI18n();
  const { pushMessage } = useToast();
  const navigate = useNavigate();
  const apiBaseURL = useAtomValue(serverURL);

  const [updates, setUpdates] = useState<SubscriptionVideoUpdate[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchUpdates = useCallback(async () => {
    if (!apiBaseURL) {
        setIsLoading(false);
        setError("Server URL not configured.");
        return;
    }
    setIsLoading(true);
    setError(null);
    try {
      // Adjust endpoint as needed, e.g., to fetch only unseen or paginate
      const response = await ffetch<SubscriptionVideoUpdate[]>(`${apiBaseURL}/api/subscriptions/updates?seen=false&limit=50`);
      if (Array.isArray(response)) {
        setUpdates(response);
      } else {
        console.error('Invalid response structure from /api/subscriptions/updates:', response);
        setUpdates([]);
        setError(i18n.t('errorFetchingUpdates', { message: 'Invalid response format' }));
      }
    } catch (err: any) {
      console.error('Failed to fetch new video updates:', err);
      setError(i18n.t('errorFetchingUpdates', { message: err.message || 'Unknown error' }));
      setUpdates([]);
    } finally {
      setIsLoading(false);
    }
  }, [apiBaseURL, i18n]);

  useEffect(() => {
    fetchUpdates();
  }, [fetchUpdates]);

  const handleMarkAsSeen = async (updateId: string) => {
    if (!apiBaseURL) return;
    try {
      await ffetch(`${apiBaseURL}/api/subscriptions/updates/${updateId}/seen`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ seen: true }),
      });
      pushMessage(i18n.t('videoMarkedAsSeenToast'), 'success');
      fetchUpdates(); // Refetch to update the list
    } catch (err: any) {
      pushMessage(i18n.t('errorMarkingAsSeen', { message: err.message }), 'error');
    }
  };

  const handleDownload = async (updateId: string, videoTitle: string) => {
    if (!apiBaseURL) return;
    try {
      await ffetch(`${apiBaseURL}/api/subscriptions/updates/${updateId}/download`, {
        method: 'POST',
      });
      pushMessage(i18n.t('downloadQueuedToast', { title: videoTitle }), 'success');
      // Optionally, update status locally or refetch
    } catch (err: any) {
      pushMessage(i18n.t('errorStartingDownload', { message: err.message }), 'error');
    }
  };

  const handleViewChannel = (subscriptionId: string) => {
    // For now, navigating to the main subscriptions page.
    // A more specific navigation could be to /subscriptions/{subscriptionId}/videos if such a route exists
    // or by passing state/query params if Subscriptions.tsx can handle it.
    navigate('/subscriptions');
    console.log("Navigate to view channel for subscription ID:", subscriptionId); // Placeholder action
  };

  if (isLoading) {
    return (
      <Container sx={{ py: 4, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress />
      </Container>
    );
  }

  if (error) {
    return (
      <Container sx={{ py: 4 }}>
        <Typography color="error" textAlign="center">{error}</Typography>
      </Container>
    );
  }

  return (
    <Container sx={{ py: 4 }}>
      <Typography variant="h4" component="h1" gutterBottom>
        {i18n.t('newVideosPageTitle') || "New Video Updates"}
      </Typography>

      {updates.length === 0 ? (
        <Typography textAlign="center">{i18n.t('noNewVideoUpdates') || "No new video updates found."}</Typography>
      ) : (
        <Grid container spacing={3}>
          {updates.map((update) => (
            <Grid item xs={12} sm={6} md={4} lg={3} key={update.id}>
              <Card sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
                {update.thumbnail_url ? (
                  <CardMedia
                    component="img"
                    height="140"
                    image={update.thumbnail_url}
                    alt={update.video_title}
                  />
                ) : (
                  <Box sx={{ height: 140, display: 'flex', alignItems: 'center', justifyContent: 'center', backgroundColor: 'grey.200' }}>
                    <MovieIcon sx={{ fontSize: 60, color: 'grey.500' }} />
                  </Box>
                )}
                <CardContent sx={{ flexGrow: 1 }}>
                  <Tooltip title={update.video_title}>
                    <Typography gutterBottom variant="subtitle1" component="div" noWrap>
                      {update.video_title}
                    </Typography>
                  </Tooltip>
                  {update.published_at && (
                     <Typography variant="body2" color="text.secondary">
                       {i18n.t('publishedAtLabel') || "Published:"} {formatDate(update.published_at.toString())} {/* Ensure formatDate handles string date */}
                     </Typography>
                   )}
                </CardContent>
                <CardActions sx={{ justifyContent: 'space-around', flexWrap: 'wrap' }}>
                  <Tooltip title={i18n.t('markAsSeenButtonLabel') || "Mark as Seen"}>
                    <IconButton onClick={() => handleMarkAsSeen(update.id)} size="small">
                      <CheckIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title={i18n.t('downloadVideoButtonLabel') || "Download"}>
                    <IconButton onClick={() => handleDownload(update.id, update.video_title)} size="small">
                      <DownloadIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title={i18n.t('viewChannelButtonLabel') || "View Channel"}>
                    <IconButton onClick={() => handleViewChannel(update.subscription_id)} size="small">
                      <PlaylistPlayIcon />
                    </IconButton>
                  </Tooltip>
                </CardActions>
              </Card>
            </Grid>
          ))}
        </Grid>
      )}
    </Container>
  );
};

export default NewVideosViewer;
