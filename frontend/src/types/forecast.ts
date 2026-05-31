export interface ForecastMonth {
  /** "YYYY-MM" */
  month: string;
  inflow: number;
  outflow: number;
  net: number;
  projected_balance: number;
}

export interface ForecastAlert {
  /** "YYYY-MM" */
  month: string;
  message: string;
}

export interface Forecast {
  /** All values in cents (int64). */
  current_balance: number;
  months: ForecastMonth[];
  end_balance: number;
  lowest: {
    month: string;
    balance: number;
  };
  alerts: ForecastAlert[];
}
