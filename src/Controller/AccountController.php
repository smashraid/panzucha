<?php

namespace App\Controller;

use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;
use Symfony\Component\Security\Core\Encoder\UserPasswordEncoderInterface;
use App\Entity\Order;
use App\Entity\OrderLine;
use App\Entity\User;
use App\Form\ChangePasswordFormType;
use App\Form\CustomerFormType;

class AccountController extends AbstractController
{
    /**
     * @Route({
     *      "en": "/account",
     *      "es": "/cuenta"
     * }, name="account")
     */
    public function index(Request $request): Response
    {
        $entityManager = $this->getDoctrine()->getManager();
        $repository = $this->getDoctrine()->getRepository(User::class);
        $user = $this->getUser();
        $form = $this->createForm(CustomerFormType::class, $user);
        $form->handleRequest($request);

        if ($form->isSubmitted() && $form->isValid()) {
            
        }

        return $this->render('account/index.html.twig', [
            'controller_name' => 'AccountController',
            'form' => $form->createView(),
            'addresses' => $user->getAddresses(),
        ]);
    }

    /**
     * @Route("/security", name="security")
     */
    public function security(Request $request, UserPasswordEncoderInterface $passwordEncoder) : Response
    {
        $form = $this->createForm(ChangePasswordFormType::class);
        $form->handleRequest($request);

        if ($form->isSubmitted() && $form->isValid()) {
            $user = $this->getUser();
            //$checkPass = $passwordEncoder->isPasswordValid($user, $old_pwd);
            $encodedPassword = $passwordEncoder->encodePassword(
                $user,
                $form->get('plainPassword')->getData()
            );

            $user->setPassword($encodedPassword);
            $this->getDoctrine()->getManager()->flush();
        }

        return $this->render('account/security.html.twig', [
            'controller_name' => 'AccountController',
            'resetForm' => $form->createView(),
        ]);
    }

    /**
     * @Route("/order", name="order")
     */
    public function order(Request $request): Response
    {
        $entityManager = $this->getDoctrine()->getManager();
        $repository = $this->getDoctrine()->getRepository(Order::class);
        $user = $this->getUser();
        $orders =  $repository->findBy(
            ['customer' => $user->getId()],
            ['orderDate' => 'DESC'],            
        );

        return $this->render('account/order.html.twig', [
            'controller_name' => 'AccountController',
            'orders' => $orders,
        ]);
    }

    /**
     * @Route("orderdetail", name="order_detail")
     */
    public function order_detail(Request $request, Order $order): Response
    {
        $entityManager = $this->getDoctrine()->getManager();
        $repository = $this->getDoctrine()->getRepository(OrderLine::class);
        $user = $this->getUser();
        $orderLines =  $repository->findBy(
            ['order1' => $order->getId()]      
        );

        return $this->render('account/order_detail.html.twig', [
            'controller_name' => 'AccountController',
            'orderLines' => $orderLines,
        ]);
    }
}
