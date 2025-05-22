import React, { useState } from 'react';
import {
  Box, Typography, Button, CircularProgress, Paper, Grid, Card, CardContent,
  CardMedia, IconButton, Collapse // Collapse isn't used in the provided code, can remove if not needed
} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'; // For show more/less
import { YtdlpChannelDump, YtdlpVideoInfo } from '../../types'; // Adjust path as needed
import { useI18n } from '../../hooks/useI18n';
import { formatDuration, formatDate } from '../../utils'; // Assuming these utils exist or can be created

interface ChannelVideosViewProps {
  channelData: YtdlpChannelDump | null;
  onClose: () => void;
  isLoading: boolean;
}

const VIDEO_LIMIT_INITIAL = 5;

const ChannelVideosView: React.FC<ChannelVideosViewProps> = ({ channelData, onClose, isLoading }) => {
  const { i18n } = useI18n();
  const [showAllVideos, setShowAllVideos] = useState(false);

  if (isLoading) {
    return (
      <Paper sx={{ p: 3, mt: 2, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress />
      </Paper>
    );
  }

  if (!channelData) {
    return null; // Or some placeholder if needed when data is cleared but component is still mounted
  }

  // Sort entries by upload_date if available, otherwise use as is.
  // yt-dlp entries for channels are often reverse chronological.
  const sortedEntries = channelData.entries ? [...channelData.entries].sort((a, b) => {
    if (a.upload_date && b.upload_date) {
      return b.upload_date.localeCompare(a.upload_date); // Newest first
    }
    if (a.playlist_index && b.playlist_index) { // For playlists
        return a.playlist_index - b.playlist_index;
    }
    return 0;
  }) : [];

  const videosToShow = showAllVideos ? sortedEntries : sortedEntries.slice(0, VIDEO_LIMIT_INITIAL);

  const handleDownloadVideo = (video: YtdlpVideoInfo) => {
    // TODO: Implement actual download logic.
    // This will involve calling an RPC method or API endpoint,
    // potentially passing video.webpage_url or video.id,
    // and user's format/quality preferences.
    // Also needs to consider the channel-specific download path.
    console.log('Download requested for:', video.title, video.webpage_url);
    // Example: rpc.exec(video.webpage_url, preferred_format_options)
  };

  return (
    <Paper sx={{ p: 2, mt: 2, mb: 4 }}>
      <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
        <Typography variant="h5">{channelData.title || i18n.t('channelVideosTitle')}</Typography>
        <Button variant="outlined" onClick={onClose}>{i18n.t('closeButtonLabel')}</Button>
      </Box>

      {sortedEntries.length === 0 && (
        <Typography>{i18n.t('noVideosFoundInChannel')}</Typography>
      )}

      <Grid container spacing={2}>
        {videosToShow.map((video) => (
          <Grid item xs={12} sm={6} md={4} lg={3} key={video.id || video.webpage_url}>
            <Card sx={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
              {video.thumbnail && (
                <CardMedia
                  component="img"
                  height="140"
                  image={video.thumbnail}
                  alt={video.title}
                />
              )}
              <CardContent sx={{ flexGrow: 1 }}>
                <Typography gutterBottom variant="subtitle1" component="div" sx={{ maxHeight: '3.6em', overflow: 'hidden', textOverflow: 'ellipsis', display: '-webkit-box', WebkitLineClamp: 2, WebkitBoxOrient: 'vertical' }}>
                  {video.title || i18n.t('untitledVideo')}
                </Typography>
                {video.uploader && <Typography variant="body2" color="text.secondary">{video.uploader}</Typography>}
                {video.upload_date && <Typography variant="body2" color="text.secondary">{formatDate(video.upload_date)}</Typography>}
                {video.duration && <Typography variant="body2" color="text.secondary">{formatDuration(video.duration)}</Typography>}
              </CardContent>
              <Box sx={{ p: 1, display: 'flex', justifyContent: 'flex-end' }}>
                <IconButton onClick={() => handleDownloadVideo(video)} title={i18n.t('downloadVideoButtonTitle')}>
                  <DownloadIcon />
                </IconButton>
              </Box>
            </Card>
          </Grid>
        ))}
      </Grid>

      {sortedEntries.length > VIDEO_LIMIT_INITIAL && (
        <Box sx={{ mt: 2, textAlign: 'center' }}>
          <Button
            onClick={() => setShowAllVideos(!showAllVideos)}
            startIcon={<ExpandMoreIcon style={{ transform: showAllVideos ? 'rotate(180deg)' : 'none' }} />}
          >
            {showAllVideos ? i18n.t('showLessButtonLabel') : i18n.t('showMoreButtonLabel')}
          </Button>
        </Box>
      )}
    </Paper>
  );
};

export default ChannelVideosView;
