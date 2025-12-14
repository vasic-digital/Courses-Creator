import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createNativeStackNavigator } from '@react-navigation/native-stack';
import CourseListScreen from './src/screens/CourseListScreen';
import CoursePlayerScreen from './src/screens/CoursePlayerScreen';

const Stack = createNativeStackNavigator();

const App: React.FC = () => {
  return (
    <NavigationContainer>
      <Stack.Navigator initialRouteName="CourseList">
        <Stack.Screen
          name="CourseList"
          component={CourseListScreen}
          options={{ title: 'Course Creator Player' }}
        />
        <Stack.Screen
          name="CoursePlayer"
          component={CoursePlayerScreen}
          options={{ title: 'Course Player' }}
        />
      </Stack.Navigator>
    </NavigationContainer>
  );
};

export default App;