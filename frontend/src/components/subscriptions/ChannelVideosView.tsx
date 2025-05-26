import React, { useState } from 'react';
import {
  Box, Typography, Button, CircularProgress, Paper, Grid, Card, CardContent,
  CardMedia, IconButton
} from '@mui/material';
import DownloadIcon from '@mui/icons-material/Download';
import CheckCircleIcon from '@mui/icons-material/CheckCircle'; 
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'; 
import { YtdlpChannelDump, YtdlpVideoInfo } from '../../types'; 
import { useI18n } from '../../hooks/useI18n';
import { useRPC } from '../../hooks/useRPC'; 
import { useToast } from '../../hooks/toast'; 
import { formatDuration, formatDate } from '../../utils'; 
// Added imports for Jotai and preference atoms
import { useAtomValue } from 'jotai';
import { preferredFormatsAtom, preferredQualitiesAtom } from '../../atoms/settings'; // PreferenceItem not directly used here

interface ChannelVideosViewProps {
  channelData: YtdlpChannelDump | null;
  onClose: () => void;
  isLoading: boolean;
  // Optional: Pass down user-defined base path and filename template if available from settings
  // userBasePath?: string; 
  // userFilenameTemplate?: string;
}

const VIDEO_LIMIT_INITIAL = 5;

const ChannelVideosView: React.FC<ChannelVideosViewProps> = ({ channelData, onClose, isLoading }) => {
  const { i18n } = useI18n();
  const { client } = useRPC(); 
  const { pushMessage } = useToast(); 
  const [showAllVideos, setShowAllVideos] = useState(false);

  // Read preference atoms
  const preferredFormats = useAtomValue(preferredFormatsAtom);
  const preferredQualities = useAtomValue(preferredQualitiesAtom);

  if (isLoading) {
    return (
      <Paper sx={{ p: 3, mt: 2, display: 'flex', justifyContent: 'center' }}>
        <CircularProgress />
      </Paper>
    );
  }

  if (!channelData) {
    return null; 
  }

  const sortedEntries = channelData.entries ? [...channelData.entries].sort((a, b) => {
    if (a.upload_date && b.upload_date) {
      return b.upload_date.localeCompare(a.upload_date); 
    }
    if (a.playlist_index && b.playlist_index) { 
        return a.playlist_index - b.playlist_index;
    }
    return 0;
  }) : [];

  const videosToShow = showAllVideos ? sortedEntries : sortedEntries.slice(0, VIDEO_LIMIT_INITIAL);

  const handleDownloadVideo = async (video: YtdlpVideoInfo) => {
    if (!video.webpage_url) {
      pushMessage(i18n.t('errorMissingVideoUrl'), 'error');
      return;
    }

    let channelFolderName = "";
    if (channelData && channelData.title) { 
      channelFolderName = channelData.title;
    } else if (video.uploader) { 
      channelFolderName = video.uploader;
    }
    
    channelFolderName = channelFolderName.replace(/[<>:"/\\|?*]+/g, '_').replace(/\.\./g, '_');
    const originalFolderValue = (channelData && channelData.title) || video.uploader || "";
    if (originalFolderValue !== "" && channelFolderName === "") { 
        channelFolderName = "_"; 
    }

    // Extract enabled formats and qualities
    const activeFormats = preferredFormats.filter(f => f.enabled).map(f => f.value);
    const activeQualities = preferredQualities.filter(q => q.enabled).map(q => q.value);

    try {
      await client.download({
        url: video.webpage_url!, 
        args: "", // No raw CLI args from here for individual video download
        channel_folder: channelFolderName || undefined, 
        playlist: false, 
        preferred_formats: activeFormats.length > 0 ? activeFormats : undefined,
        preferred_qualities: activeQualities.length > 0 ? activeQualities : undefined,
      });
      
      pushMessage(i18n.t('downloadStartedSuccess', { title: video.title }), 'success');
    } catch (error: any) {
      console.error('Failed to start download:', error);
      pushMessage(i18n.t('errorStartingDownload', { message: error.message || 'Unknown error' }), 'error');
    }
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
                  {video.is_downloaded && <CheckCircleIcon color="success" sx={{ fontSize: '1rem', ml: 0.5, verticalAlign: 'middle' }} />}
                </Typography>
                {video.uploader && <Typography variant="body2" color="text.secondary">{video.uploader}</Typography>}
                {video.upload_date && <Typography variant="body2" color="text.secondary">{formatDate(video.upload_date)}</Typography>}
                {video.duration && <Typography variant="body2" color="text.secondary">{formatDuration(video.duration)}</Typography>}

                {/* Display Playlist Info */}
                {video.playlist_title && video.playlist_index && (
                  <Typography variant="body2" color="text.secondary">
                    {i18n.t('playlistInfoLabel', { title: video.playlist_title, index: video.playlist_index })}
                  </Typography>
                )}

                {/* Display Series Info - Placeholder as fields are not in YtdlpVideoInfo */}
                {(video as any).season_number && (video as any).episode_number && (
                  <Typography variant="body2" color="text.secondary">
                    {i18n.t('seriesInfoLabel', { seasonNum: (video as any).season_number, episodeNum: (video as any).episode_number })}
                  </Typography>
                )}
                {(video as any).episode_title && (video as any).episode_title !== video.title && (
                  <Typography variant="body2" color="text.secondary">
                    {i18n.t('episodeInfoLabel', { title: (video as any).episode_title })}
                  </Typography>
                )}

                {/* Display Resolution - Placeholder as field is not in YtdlpVideoInfo */}
                {(video as any).resolution && (
                  <Typography variant="body2" color="text.secondary">
                    {i18n.t('resolutionLabel', { resolution: (video as any).resolution })}
                  </Typography>
                )}
              </CardContent>
              <Box sx={{ p: 1, display: 'flex', justifyContent: 'flex-end' }}>
                <IconButton
                  onClick={() => handleDownloadVideo(video)}
                  title={video.is_downloaded ? i18n.t('videoAlreadyDownloaded') : i18n.t('downloadVideoButtonTitle')}
                  disabled={video.is_downloaded}
                >
                  {video.is_downloaded ? <CheckCircleIcon color="success" /> : <DownloadIcon />}
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
