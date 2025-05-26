import React from 'react';
import { Card, CardActionArea, CardMedia, CardContent, Typography, Box } from '@mui/material';
import { ArchiveEntry } from '../../types'; // Adjust path if needed
import { useI18n } from '../../hooks/useI18n';
import { formatDuration } from '../../utils'; // Assuming created in Subtask 3.B.2
// Import an icon for missing thumbnail, e.g., MovieIcon or BrokenImageIcon
import MovieIcon from '@mui/icons-material/Movie'; 

interface MediaCardProps {
  entry: ArchiveEntry;
  onClick?: (entry: ArchiveEntry) => void; // For opening detailed view later
}

const MediaCard: React.FC<MediaCardProps> = ({ entry, onClick }) => {
  const { i18n } = useI18n();
  let metadata: any = {};
  try {
    if (entry.metadata) {
      metadata = JSON.parse(entry.metadata);
    }
  } catch (e) {
    console.error("Failed to parse metadata for entry:", entry.id, e);
  }

  const title = entry.title || metadata.title || i18n.t('untitledMedia');
  const thumbnailUrl = entry.thumbnail || metadata.thumbnail;
  const duration = metadata.duration ? formatDuration(metadata.duration) : null;
  const uploader = metadata.uploader || metadata.channel || (entry.source && new URL(entry.source).hostname); // Basic fallback for uploader
  // Add more fields like resolution if available

  const handleCardClick = () => {
    if (onClick) {
      onClick(entry);
    } else {
      // Default action if no onClick provided, e.g., log or nothing
      console.log("Media card clicked:", entry.id);
    }
  };

  return (
    <Card sx={{ height: '100%', display: 'flex', flexDirection: 'column' }}>
      <CardActionArea onClick={handleCardClick} sx={{ flexGrow: 1, display: 'flex', flexDirection: 'column' }}>
        <Box sx={{ position: 'relative', width: '100%', paddingTop: '56.25%' /* 16:9 Aspect Ratio */ }}>
          {thumbnailUrl ? (
            <CardMedia
              component="img"
              image={thumbnailUrl}
              alt={title}
              sx={{ position: 'absolute', top: 0, left: 0, width: '100%', height: '100%', objectFit: 'cover' }}
            />
          ) : (
            <Box sx={{ position: 'absolute', top: 0, left: 0, width: '100%', height: '100%', display: 'flex', alignItems: 'center', justifyContent: 'center', backgroundColor: 'grey.200' }}>
              <MovieIcon sx={{ fontSize: 60, color: 'grey.500' }} />
            </Box>
          )}
          {duration && (
            <Typography
              variant="caption"
              sx={{
                position: 'absolute',
                bottom: 8,
                right: 8,
                backgroundColor: 'rgba(0, 0, 0, 0.7)',
                color: 'white',
                padding: '2px 6px',
                borderRadius: '4px',
              }}
            >
              {duration}
            </Typography>
          )}
        </Box>
        <CardContent sx={{ width: '100%', flexGrow: 1 }}>
          <Typography gutterBottom variant="subtitle1" component="div" title={title} sx={{
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              display: '-webkit-box',
              WebkitLineClamp: 2, // Show max 2 lines
              WebkitBoxOrient: 'vertical',
          }}>
            {title}
          </Typography>
          {uploader && (
            <Typography variant="body2" color="text.secondary" noWrap title={uploader}>
              {uploader}
            </Typography>
          )}
          {/* Display Playlist Info */}
          {metadata.playlist_title && metadata.playlist_index && (
            <Typography variant="body2" color="text.secondary" noWrap title={`${metadata.playlist_title} (Video ${metadata.playlist_index})`}>
              {i18n.t('playlistInfoLabel', { title: metadata.playlist_title, index: metadata.playlist_index })}
            </Typography>
          )}
          {/* Display Series Info - using common yt-dlp field names */}
          {metadata.season_number && metadata.episode_number && (
            <Typography variant="body2" color="text.secondary" noWrap>
              {i18n.t('seriesInfoLabel', { seasonNum: metadata.season_number, episodeNum: metadata.episode_number })}
            </Typography>
          )}
          {/* Display Episode Title if different from main title and series info is not present (to avoid redundancy) */}
          {metadata.episode_title && metadata.episode_title !== title && !metadata.season_number && (
            <Typography variant="body2" color="text.secondary" noWrap title={metadata.episode_title}>
              {i18n.t('episodeInfoLabel', { title: metadata.episode_title })}
            </Typography>
          )}
          {/* Display Resolution */}
          {(metadata.resolution || (metadata.width && metadata.height)) && (
            <Typography variant="body2" color="text.secondary" noWrap>
              {i18n.t('resolutionLabel', { resolution: metadata.resolution || `${metadata.width}x${metadata.height}` })}
            </Typography>
          )}
        </CardContent>
      </CardActionArea>
    </Card>
  );
};

export default MediaCard;
