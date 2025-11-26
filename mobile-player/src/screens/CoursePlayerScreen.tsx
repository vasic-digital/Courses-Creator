import React, { useState, useRef } from 'react';
import {
  View,
  Text,
  TouchableOpacity,
  StyleSheet,
  ScrollView,
  Dimensions,
} from 'react-native';
import { Course, Lesson, PlayerState } from '../types';

const { width } = Dimensions.get('window');

interface CoursePlayerScreenProps {
  route: any;
  navigation: any;
}

const CoursePlayerScreen: React.FC<CoursePlayerScreenProps> = ({ route, navigation }) => {
  const { course }: { course: Course } = route.params;
  const [currentLessonIndex, setCurrentLessonIndex] = useState(0);
  const [isPlaying, setIsPlaying] = useState(false);
  const [playerState, setPlayerState] = useState<PlayerState>({
    currentLesson: 0,
    isPlaying: false,
    currentTime: 0,
    volume: 1,
    subtitlesEnabled: false,
  });

  const currentLesson = course.lessons[currentLessonIndex];

  const togglePlayPause = () => {
    setIsPlaying(!isPlaying);
    setPlayerState(prev => ({ ...prev, isPlaying: !prev.isPlaying }));
  };

  const nextLesson = () => {
    if (currentLessonIndex < course.lessons.length - 1) {
      setCurrentLessonIndex(currentLessonIndex + 1);
      setPlayerState(prev => ({ ...prev, currentLesson: currentLessonIndex + 1 }));
    }
  };

  const previousLesson = () => {
    if (currentLessonIndex > 0) {
      setCurrentLessonIndex(currentLessonIndex - 1);
      setPlayerState(prev => ({ ...prev, currentLesson: currentLessonIndex - 1 }));
    }
  };

  const renderLesson = (lesson: Lesson, index: number) => (
    <TouchableOpacity
      key={lesson.id}
      style={[
        styles.lessonItem,
        index === currentLessonIndex && styles.currentLesson,
      ]}
      onPress={() => setCurrentLessonIndex(index)}
    >
      <Text style={styles.lessonNumber}>{index + 1}</Text>
      <View style={styles.lessonContent}>
        <Text style={styles.lessonTitle}>{lesson.title}</Text>
        <Text style={styles.lessonDuration}>{Math.floor(lesson.duration / 60)}:{(lesson.duration % 60).toString().padStart(2, '0')}</Text>
      </View>
    </TouchableOpacity>
  );

  return (
    <View style={styles.container}>
      {/* Video Player Area */}
      <View style={styles.playerContainer}>
        <View style={styles.videoPlaceholder}>
          <Text style={styles.placeholderText}>
            {currentLesson?.videoUrl ? 'Video Player' : 'Audio Player'}
          </Text>
          <Text style={styles.lessonTitleText}>{currentLesson?.title}</Text>
        </View>

        {/* Player Controls */}
        <View style={styles.controls}>
          <TouchableOpacity onPress={previousLesson} disabled={currentLessonIndex === 0}>
            <Text style={[styles.controlButton, currentLessonIndex === 0 && styles.disabledButton]}>⏮️</Text>
          </TouchableOpacity>

          <TouchableOpacity onPress={togglePlayPause}>
            <Text style={styles.controlButton}>{isPlaying ? '⏸️' : '▶️'}</Text>
          </TouchableOpacity>

          <TouchableOpacity onPress={nextLesson} disabled={currentLessonIndex === course.lessons.length - 1}>
            <Text style={[styles.controlButton, currentLessonIndex === course.lessons.length - 1 && styles.disabledButton]}>⏭️</Text>
          </TouchableOpacity>
        </View>
      </View>

      {/* Lesson Content */}
      <View style={styles.contentContainer}>
        <Text style={styles.sectionTitle}>Course Content</Text>
        <ScrollView style={styles.lessonsList}>
          {course.lessons.map((lesson, index) => renderLesson(lesson, index))}
        </ScrollView>
      </View>
    </View>
  );
};

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f5f5f5',
  },
  playerContainer: {
    backgroundColor: '#000',
    height: 250,
  },
  videoPlaceholder: {
    flex: 1,
    justifyContent: 'center',
    alignItems: 'center',
    backgroundColor: '#333',
  },
  placeholderText: {
    color: '#fff',
    fontSize: 18,
    marginBottom: 10,
  },
  lessonTitleText: {
    color: '#fff',
    fontSize: 16,
    textAlign: 'center',
  },
  controls: {
    flexDirection: 'row',
    justifyContent: 'center',
    alignItems: 'center',
    padding: 10,
    backgroundColor: '#222',
  },
  controlButton: {
    fontSize: 24,
    marginHorizontal: 20,
    color: '#fff',
  },
  disabledButton: {
    opacity: 0.5,
  },
  contentContainer: {
    flex: 1,
    padding: 16,
  },
  sectionTitle: {
    fontSize: 20,
    fontWeight: 'bold',
    marginBottom: 16,
    color: '#333',
  },
  lessonsList: {
    flex: 1,
  },
  lessonItem: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: 12,
    marginBottom: 8,
    backgroundColor: 'white',
    borderRadius: 8,
  },
  currentLesson: {
    backgroundColor: '#e3f2fd',
    borderColor: '#2196f3',
    borderWidth: 2,
  },
  lessonNumber: {
    width: 30,
    height: 30,
    borderRadius: 15,
    backgroundColor: '#2196f3',
    color: 'white',
    textAlign: 'center',
    textAlignVertical: 'center',
    fontWeight: 'bold',
    marginRight: 12,
  },
  lessonContent: {
    flex: 1,
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  lessonTitle: {
    fontSize: 16,
    color: '#333',
    flex: 1,
  },
  lessonDuration: {
    fontSize: 14,
    color: '#666',
  },
});

export default CoursePlayerScreen;