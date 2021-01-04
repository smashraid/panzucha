<?php

namespace App\Entity;

use App\Repository\OrderRepository;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;
use Doctrine\ORM\Mapping as ORM;

/**
 * @ORM\Entity(repositoryClass=OrderRepository::class)
 * @ORM\Table(name="`order`")
 */
class Order
{
    /**
     * @ORM\Id
     * @ORM\GeneratedValue
     * @ORM\Column(type="integer")
     */
    private $id;

    /**
     * @ORM\Column(type="string", length=20)
     */
    private $status;

    /**
     * @ORM\Column(type="datetime")
     */
    private $orderDate;

    /**
     * @ORM\Column(type="float")
     */
    private $shippingAmount;

    /**
     * @ORM\Column(type="float")
     */
    private $priceAmount;

    /**
     * @ORM\Column(type="float")
     */
    private $orderSubtotal;

    /**
     * @ORM\Column(type="string", length=255)
     */
    private $shippingStreet1;

    /**
     * @ORM\Column(type="string", length=255, nullable=true)
     */
    private $shippingStreet2;

    /**
     * @ORM\Column(type="string", length=255)
     */
    private $shippingCity;

    /**
     * @ORM\Column(type="string", length=100)
     */
    private $shippingState;

    /**
     * @ORM\Column(type="string", length=100)
     */
    private $shippingCountry;

    /**
     * @ORM\Column(type="string", length=10)
     */
    private $shippingPostalCode;

    /**
     * @ORM\Column(type="string", length=255, nullable=true)
     */
    private $specialInstructions;

    /**
     * @ORM\Column(type="string", length=255, nullable=true)
     */
    private $transactionId;

    /**
     * @ORM\OneToMany(targetEntity=OrderLine::class, mappedBy="order1", orphanRemoval=true)
     */
    private $orderLines;

    /**
     * @ORM\ManyToOne(targetEntity=User::class, inversedBy="orders")
     * @ORM\JoinColumn(nullable=false)
     */
    private $customer;

    public function __construct()
    {
        $this->orderLines = new ArrayCollection();
    }

    public function getId(): ?int
    {
        return $this->id;
    }

    public function getStatus(): ?string
    {
        return $this->status;
    }

    public function setStatus(string $status): self
    {
        $this->status = $status;

        return $this;
    }

    public function getOrderDate(): ?\DateTimeInterface
    {
        return $this->orderDate;
    }

    public function setOrderDate(\DateTimeInterface $orderDate): self
    {
        $this->orderDate = $orderDate;

        return $this;
    }

    public function getShippingAmount(): ?float
    {
        return $this->shippingAmount;
    }

    public function setShippingAmount(float $shippingAmount): self
    {
        $this->shippingAmount = $shippingAmount;

        return $this;
    }

    public function getPriceAmount(): ?float
    {
        return $this->priceAmount;
    }

    public function setPriceAmount(float $priceAmount): self
    {
        $this->priceAmount = $priceAmount;

        return $this;
    }

    public function getOrderSubtotal(): ?float
    {
        return $this->orderSubtotal;
    }

    public function setOrderSubtotal(float $orderSubtotal): self
    {
        $this->orderSubtotal = $orderSubtotal;

        return $this;
    }

    public function getShippingStreet1(): ?string
    {
        return $this->shippingStreet1;
    }

    public function setShippingStreet1(string $shippingStreet1): self
    {
        $this->shippingStreet1 = $shippingStreet1;

        return $this;
    }

    public function getShippingStreet2(): ?string
    {
        return $this->shippingStreet2;
    }

    public function setShippingStreet2(?string $shippingStreet2): self
    {
        $this->shippingStreet2 = $shippingStreet2;

        return $this;
    }

    public function getShippingCity(): ?string
    {
        return $this->shippingCity;
    }

    public function setShippingCity(string $shippingCity): self
    {
        $this->shippingCity = $shippingCity;

        return $this;
    }

    public function getShippingState(): ?string
    {
        return $this->shippingState;
    }

    public function setShippingState(string $shippingState): self
    {
        $this->shippingState = $shippingState;

        return $this;
    }

    public function getShippingCountry(): ?string
    {
        return $this->shippingCountry;
    }

    public function setShippingCountry(string $shippingCountry): self
    {
        $this->shippingCountry = $shippingCountry;

        return $this;
    }

    public function getShippingPostalCode(): ?string
    {
        return $this->shippingPostalCode;
    }

    public function setShippingPostalCode(string $shippingPostalCode): self
    {
        $this->shippingPostalCode = $shippingPostalCode;

        return $this;
    }

    public function getSpecialInstructions(): ?string
    {
        return $this->specialInstructions;
    }

    public function setSpecialInstructions(?string $specialInstructions): self
    {
        $this->specialInstructions = $specialInstructions;

        return $this;
    }

    public function getTransactionId(): ?string
    {
        return $this->transactionId;
    }

    public function setTransactionId(?string $transactionId): self
    {
        $this->transactionId = $transactionId;

        return $this;
    }

    /**
     * @return Collection|OrderLine[]
     */
    public function getOrderLines(): Collection
    {
        return $this->orderLines;
    }

    public function addOrderLine(OrderLine $orderLine): self
    {
        if (!$this->orderLines->contains($orderLine)) {
            $this->orderLines[] = $orderLine;
            $orderLine->setOrder1($this);
        }

        return $this;
    }

    public function removeOrderLine(OrderLine $orderLine): self
    {
        if ($this->orderLines->removeElement($orderLine)) {
            // set the owning side to null (unless already changed)
            if ($orderLine->getOrder1() === $this) {
                $orderLine->setOrder1(null);
            }
        }

        return $this;
    }

    public function getCustomer(): ?User
    {
        return $this->customer;
    }

    public function setCustomer(?User $customer): self
    {
        $this->customer = $customer;

        return $this;
    }
}
