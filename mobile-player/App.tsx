import React, { Suspense } from "react";
import { NavigationContainer } from "@react-navigation/native";
import { createNativeStackNavigator } from "@react-navigation/native-stack";
import { ActivityIndicator, View } from "react-native";

const CourseListScreen = React.lazy(() => import("./src/screens/CourseListScreen"));
const CoursePlayerScreen = React.lazy(() => import("./src/screens/CoursePlayerScreen"));

const Stack = createNativeStackNavigator();

const App: React.FC = () => {
  return (
    <NavigationContainer>
      <Suspense fallback={<View style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}><ActivityIndicator size="large" /></View>}>
        <Stack.Navigator initialRouteName="CourseList">
          <Stack.Screen
            name="CourseList"
            component={CourseListScreen}
            options={{ title: "Course Creator Player" }}
          />
          <Stack.Screen
            name="CoursePlayer"
            component={CoursePlayerScreen}
            options={{ title: "Course Player" }}
          />
        </Stack.Navigator>
      </Suspense>
    </NavigationContainer>
  );
};

export default App;
