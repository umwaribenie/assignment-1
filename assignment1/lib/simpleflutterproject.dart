import 'package:flutter/material.dart';

void main() {
  runApp(simplecalculator());
}

class simplecalculator extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(
          title: Text('Simple Flutter Project'),
        ),
        body: Center(
          child: Text(
            'Hey. My name is Umwari Benie',
            style: TextStyle(fontSize: 50.0,color: Colors.red),
          ),
        ),
      ),
    );
  }
}
