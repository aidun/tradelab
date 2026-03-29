import React from "react";
import { render, screen } from "@testing-library/react";
import { Hero } from "@/components/hero";

describe("Hero", () => {
  it("renders the product heading and the XRP reference market", () => {
    render(<Hero />);

    expect(screen.getByRole("heading", { name: /paper trading with/i })).toBeInTheDocument();
    expect(screen.getByText(/xrp \/ usdt/i)).toBeInTheDocument();
  });
});
