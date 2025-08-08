import { ethers } from "ethers";
import dotenv from "dotenv";
dotenv.config();

export const provider = new ethers.JsonRpcProvider(process.env.RPC_URL);
export const wallet = new ethers.Wallet(process.env.PRIVATE_KEY!, provider);
