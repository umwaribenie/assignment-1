# Flutter Counter App - Complete Beginner's Guide

A simple Flutter counter app that demonstrates basic Flutter concepts. This app displays a counter that increases when you tap the floating action button.

## What This App Does
- Shows a counter starting at 0
- Has a button (+ icon) that increases the counter when pressed
- Demonstrates basic Flutter widgets and state management

## Prerequisites (What You Need Before Starting)
- A computer running Linux, macOS, or Windows
- Internet connection for downloading Flutter
- Basic familiarity with using terminal/command line

## Step-by-Step Setup Guide

### Step 1: Install Flutter

#### For Linux (Your Current System):
1. **Download Flutter:**
   ```bash
   cd ~
   wget https://storage.googleapis.com/flutter_infra_release/releases/stable/linux/flutter_linux_3.16.0-stable.tar.xz
   ```

2. **Extract Flutter:**
   ```bash
   tar xf flutter_linux_3.16.0-stable.tar.xz
   ```

3. **Add Flutter to your PATH:**
   ```bash
   echo 'export PATH="$HOME/flutter/bin:$PATH"' >> ~/.bashrc
   source ~/.bashrc
   ```

4. **Verify installation:**
   ```bash
   flutter --version
   ```

### Step 2: Install Required Dependencies
Run the Flutter doctor to see what dependencies you need:
```bash
flutter doctor
```

Install missing dependencies as suggested by flutter doctor.

### Step 3: Enable Web Support (Easiest for Beginners)
```bash
flutter config --enable-web
```

## How to Run This App

### Method 1: Run on Web (Recommended for Beginners)
1. **Navigate to the project directory:**
   ```bash
   cd /workspace/assignment1
   ```

2. **Get the project dependencies:**
   ```bash
   flutter pub get
   ```

3. **Run the app in web browser:**
   ```bash
   flutter run -d web-server --web-port 8080
   ```

4. **Open your browser and go to:**
   ```
   http://localhost:8080
   ```

### Method 2: Run on Android Emulator (Advanced)
1. Install Android Studio
2. Set up an Android Virtual Device (AVD)
3. Start the emulator
4. Run: `flutter run`

### Method 3: Run on Physical Device
1. Enable Developer Options on your Android device
2. Enable USB Debugging
3. Connect device via USB
4. Run: `flutter run`

## Troubleshooting Common Issues

### "flutter: command not found"
- Make sure you added Flutter to your PATH correctly
- Restart your terminal
- Run: `source ~/.bashrc`

### "No connected devices"
- For web: Run `flutter config --enable-web`
- For mobile: Connect a device or start an emulator

### Dependencies issues
- Run: `flutter clean` then `flutter pub get`
- Check that your internet connection is working

### Port already in use
- Use a different port: `flutter run -d web-server --web-port 8081`

## Understanding the Code

### Main Components:
- **main.dart**: The entry point of the app
- **MyApp**: The root widget that sets up the app theme
- **MyHomePage**: The main screen with the counter
- **_counter**: Variable that stores the current count
- **_incrementCounter()**: Function that increases the counter

### Key Flutter Concepts Demonstrated:
- **StatefulWidget**: Widgets that can change over time
- **setState()**: Method to update the UI when data changes
- **MaterialApp**: Provides Material Design styling
- **Scaffold**: Basic page structure with app bar and body
- **FloatingActionButton**: The circular + button

## Next Steps for Learning
1. Try changing the counter to decrease instead of increase
2. Add a reset button
3. Change the app colors and theme
4. Add more widgets to the screen

## Resources for Further Learning
- [Flutter Documentation](https://docs.flutter.dev/)
- [Flutter Codelabs](https://docs.flutter.dev/codelabs)
- [Dart Language Tour](https://dart.dev/guides/language/language-tour)
- [Flutter Widget Catalog](https://docs.flutter.dev/development/ui/widgets)
