//#include <Serial.h>

// The constant that defines how long to debounce the signals.  In micro-seconds.
const long DEBOUNCE_TIME_MS = 20;

const unsigned long OUTPUT_REPORT_MIN_INTERVAL_MS = 1000;

const long SERIAL_NUMBER = 0;

unsigned int _output_state = 0;
unsigned int _input_state = 0;

void setup() {

  /* Setup the outputs and set them to 0 / off. */
  pinMode(8, OUTPUT);
  pinMode(9, OUTPUT);
  pinMode(10, OUTPUT);
  pinMode(11, OUTPUT);
  pinMode(12, OUTPUT);
  pinMode(13, OUTPUT);

  SetRelayState(0);

  /* Setup the inputs. */
  pinMode(7, INPUT_PULLUP);
  pinMode(6, INPUT_PULLUP);
  pinMode(5, INPUT_PULLUP);
  pinMode(4, INPUT_PULLUP);
  pinMode(3, INPUT_PULLUP);
  pinMode(2, INPUT_PULLUP);


  /* Start the serial port */
  Serial.begin(115200) ;
  //while (!Serial) ;
}

void loop() {
  if (true == UpdateInputState()) {
    OutputData();
  }

  while (Serial.available() > 0) {
    int cmd = Serial.read();

    switch (cmd) {
      case '?':
        Serial.println(F("s [0-63] sets the relay output bitmask\ng gets the latest output\ndata format: %02X|%02X|%02X\\n, sn, input, output\n"));
        break;
      case 's':
        _output_state = Serial.parseInt();
        SetRelayState(_output_state);
        break;
      case 'g':
        OutputData();
        break;

    }
  }
}

void OutputData()
{
  // Always send 9 characters
  // "00|00|00\n"  SN|input|output
  if( SERIAL_NUMBER < 0x10 ) {
    Serial.print(F("0"));
  }
  Serial.print(SERIAL_NUMBER, HEX);
  Serial.print(F("|"));
  if( _input_state < 0x10 ) {
    Serial.print(F("0"));
  }
  Serial.print(_input_state, HEX);
  Serial.print(F("|"));
  if( _output_state < 0x10 ) {
    Serial.print(F("0"));
  }
  Serial.print(_output_state, HEX);
  Serial.print(F("\n"));
  Serial.flush();
}

void SetRelayState( long out )
{
  int j = 8;
  if ( 0 <= out && out < 64 ) {
    for (int i = 0; i < 6; i++, j++) {
      if ((1 << i) & out) {
        digitalWrite(j, HIGH);
      } else {
        digitalWrite(j, LOW);
      }
    }
  } else {
    Serial.print( "Invalid range.  Expecting: [0-63]\n" );
  }
}

// Figure out the state of the inputs and the clock.
// Return if the output report should be sent.
bool UpdateInputState()
{
  static long debounce_time[8] = { 0L, 0L, 0L, 0L, 0L, 0L, 0L, 0L };
  static unsigned long last_report_time = 0;
  bool rv = false;

  unsigned long now = millis();

  for (int i = 2; i < 8; i++) {
    int io = digitalRead(i);
    if ( (io != bitRead(_input_state, i)) && (now > debounce_time[i]) ) {
      debounce_time[i] = now + DEBOUNCE_TIME_MS;
      if (LOW == io) {
        bitClear(_input_state, i);
      } else {
        bitSet(_input_state, i);
      }
      rv = true;
    }
  }

  // Always send a report at the minimum interval.
  if (now - last_report_time >= OUTPUT_REPORT_MIN_INTERVAL_MS ) {
    last_report_time = now;
    rv  = true;
  }

  return rv;
}

