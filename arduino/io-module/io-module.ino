//#include <Serial.h>

// The constant that defines how long to debounce the signals.  In micro-seconds.
const long DEBOUNCE_TIME_MS = 20;

const long OUTPUT_REPORT_MIN_INTERVAL_S = 1;

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
  // Always send 6 characters
  // 0|f|d\n
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
  static unsigned long _seconds_since_boot = 0;
  static long debounce_time[6] = { 0L, 0L, 0L, 0L, 0L, 0L };
  static unsigned long last_second = 0;
  static long last_report_s = 0;
  bool rv = false;

  unsigned long now = millis();
  unsigned long ms = now - last_second;

  if (ms > 1000UL) {
    last_second = now;
    _seconds_since_boot++;
    ms -= 1000UL;
  }

  for (int i = 0; i < 6; i++) {
    int io = digitalRead(i + 2);
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
  if (last_report_s + OUTPUT_REPORT_MIN_INTERVAL_S <= _seconds_since_boot ) {
    last_report_s = _seconds_since_boot;
    rv  = true;
  }

  return rv;
}

