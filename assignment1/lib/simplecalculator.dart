import 'package:flutter/material.dart';

void main() {
  runApp(CalculatorApp());
}

class CalculatorApp extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Calculator',
      theme: ThemeData(
        primarySwatch: Colors.blue,
        visualDensity: VisualDensity.adaptivePlatformDensity,
      ),
      home: CalculatorHomePage(),
    );
  }
}

class CalculatorHomePage extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('Calculator'),
      ),
      body: Container(
        color: Colors.black,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: <Widget>[
            Expanded(
              child: Container(
                padding: EdgeInsets.all(16.0),
                alignment: Alignment.bottomRight,
                child: Text(
                  '0',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 60.0,
                  ),
                ),
              ),
            ),
            buildButtonRow(['C', '( )', '%', '÷'], Colors.orange),
            buildButtonRow(['7', '8', '9', '×'], Colors.grey[800]!),
            buildButtonRow(['4', '5', '6', '-'], Colors.grey[800]!),
            buildButtonRow(['1', '2', '3', '+'], Colors.grey[800]!),
            buildButtonRow(['+/-', '0', '.','='], Colors.grey[800]!),
          ],
        ),
      ),
    );
  }

  Widget buildButtonRow(List<String> buttons, Color color) {
    return Expanded(
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: buttons.map((buttonText) {
          return Expanded(
            child: Container(
              margin: EdgeInsets.all(10.0),
              decoration: BoxDecoration(
                color: color,
                borderRadius: BorderRadius.circular(30.0),
              ),
              child: TextButton(
                onPressed: () {},
                child: Text(
                  buttonText,
                  style: TextStyle(
                    fontSize: 35.0,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }
}
