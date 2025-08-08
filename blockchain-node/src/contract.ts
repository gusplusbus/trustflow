import { wallet } from "./provider";
import ABI from "../contracts/contract.abi.json";
import dotenv from "dotenv";
dotenv.config();

const CONTRACT_ADDRESS = process.env.CONTRACT_ADDRESS!;

export const contract = new ethers.Contract(CONTRACT_ADDRESS, ABI, wallet);
